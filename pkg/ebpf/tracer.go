package ebpf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/grafana/ebpf-autoinstrument/pkg/pipe/global"

	"golang.org/x/sys/unix"

	"github.com/grafana/ebpf-autoinstrument/pkg/ebpf/goruntime"
	"github.com/grafana/ebpf-autoinstrument/pkg/ebpf/grpc"
	"github.com/grafana/ebpf-autoinstrument/pkg/ebpf/httpfltr"

	ebpfcommon "github.com/grafana/ebpf-autoinstrument/pkg/ebpf/common"

	"github.com/cilium/ebpf/rlimit"

	"github.com/cilium/ebpf/link"

	"github.com/cilium/ebpf"
	"github.com/grafana/ebpf-autoinstrument/pkg/exec"
	"github.com/grafana/ebpf-autoinstrument/pkg/goexec"

	"github.com/grafana/ebpf-autoinstrument/pkg/ebpf/nethttp"
	"github.com/mariomac/pipes/pkg/node"
	"golang.org/x/exp/slog"
)

// Tracer is an individual eBPF program (e.g. the net/http or the grpc tracers)
type Tracer interface {
	// Load the bpf object that is generated by the bpf2go compiler
	Load() (*ebpf.CollectionSpec, error)
	// Constants returns a map of constants to be overriden into the eBPF program.
	// The key is the constant name and the value is the value to overwrite.
	Constants(*exec.FileInfo, *goexec.Offsets) map[string]any
	// BpfObjects that are created by the bpf2go compiler
	BpfObjects() any
	// GoProbes returns a map with the name of Go functions that need to be inspected
	// in the executable, as well as the eBPF programs that optionally need to be
	// inserted as the Go function start and end probes
	GoProbes() map[string]ebpfcommon.FunctionPrograms
	// KProbes returns a map with the name of the kernel probes that need to be
	// tapped into. Start matches kprobe, End matches kretprobe
	KProbes() map[string]ebpfcommon.FunctionPrograms
	// Socket filters returns a list of programs that need to be loaded as a
	// generic eBPF socket filter
	SocketFilters() []*ebpf.Program
	// Run will do the action of listening for eBPF traces and forward them
	// periodically to the output channel.
	Run(context.Context, chan<- []any)
	// AddCloser adds io.Closer instances that need to be invoked when the
	// Run function ends.
	AddCloser(c ...io.Closer)
}

// TracerProvider returns a StartFuncCtx for each discovered eBPF traceable source: GRPC, HTTP...
func TracerProvider(ctx context.Context, cfg ebpfcommon.TracerConfig) ([]node.StartFuncCtx[[]any], error) { //nolint:all
	var log = logger()

	// Each program is an eBPF source: net/http, grpc...
	programs := []Tracer{
		&nethttp.Tracer{Cfg: &cfg},
		&nethttp.GinTracer{Tracer: nethttp.Tracer{Cfg: &cfg}},
		&grpc.Tracer{Cfg: &cfg},
		&goruntime.Tracer{Cfg: &cfg},
	}

	// merging all the functions from all the programs, in order to do
	// a complete inspection of the target executable
	allFuncs := allFunctionNames(programs)
	elfInfo, goffsets, err := inspect(ctx, &cfg, allFuncs)
	if err != nil {
		return nil, fmt.Errorf("inspecting offsets: %w", err)
	}

	if goffsets != nil {
		programs = filterNotFoundPrograms(programs, goffsets)
		if len(programs) == 0 {
			return nil, errors.New("no instrumentable function found")
		}
	} else {
		// We are not instrumenting a Go application, we override the programs
		// list with the generic kernel/socket space filters
		programs = []Tracer{&httpfltr.Tracer{Cfg: &cfg}}
	}

	// Instead of the executable file in the disk, we pass the /proc/<pid>/exec
	// to allow loading it from different container/pods in containerized environments
	exe, err := link.OpenExecutable(elfInfo.ProExeLinkPath)
	if err != nil {
		return nil, fmt.Errorf("opening %q executable file: %w", elfInfo.ProExeLinkPath, err)
	}
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("removing memory lock: %w", err)
	}

	pinPath, err := mountBpfPinPath(&cfg)
	if err != nil {
		return nil, fmt.Errorf("mounting BPF FS in %q: %w", cfg.BpfBaseDir, err)
	}

	if cfg.SystemWide {
		log.Info("system wide instrumentation")
	}

	// startNodes contains the eBPF programs (HTTP, GRPC tracers...) plus a function
	// that just waits for the passed context to finish before closing the BPF pin
	// path
	startNodes := []node.StartFuncCtx[[]any]{
		waitToCloseBbfPinPath(pinPath),
	}

	for _, p := range programs {
		plog := log.With("program", reflect.TypeOf(p))
		plog.Debug("loading eBPF program")
		spec, err := p.Load()
		if err != nil {
			unmountBpfPinPath(pinPath)
			return nil, fmt.Errorf("loading eBPF program: %w", err)
		}
		if err := spec.RewriteConstants(p.Constants(elfInfo, goffsets)); err != nil {
			return nil, fmt.Errorf("rewriting BPF constants definition: %w", err)
		}
		if err := spec.LoadAndAssign(p.BpfObjects(), &ebpf.CollectionOptions{
			Maps: ebpf.MapOptions{
				PinPath: pinPath,
			}}); err != nil {
			printVerifierErrorInfo(err)
			unmountBpfPinPath(pinPath)
			return nil, fmt.Errorf("loading and assigning BPF objects: %w", err)
		}
		i := instrumenter{
			exe:     exe,
			offsets: goffsets,
		}

		//Go style Uprobes
		if err := i.goprobes(p); err != nil {
			printVerifierErrorInfo(err)
			unmountBpfPinPath(pinPath)
			return nil, err
		}

		//Kprobes to be used for native instrumentation points
		if err := i.kprobes(p); err != nil {
			printVerifierErrorInfo(err)
			unmountBpfPinPath(pinPath)
			return nil, err
		}

		//Sock filters support
		if err := i.sockfilters(p); err != nil {
			printVerifierErrorInfo(err)
			unmountBpfPinPath(pinPath)
			return nil, err
		}

		startNodes = append(startNodes, p.Run)
	}

	return startNodes, nil
}

func mountBpfPinPath(cfg *ebpfcommon.TracerConfig) (string, error) {
	pinPath := path.Join(cfg.BpfBaseDir, strconv.Itoa(os.Getpid()))
	log := slog.With("component", "ebpf.TracerProvider", "path", pinPath)
	log.Debug("mounting BPF map pinning path")
	if _, err := os.Stat(pinPath); err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("accessing %s stat: %w", pinPath, err)
		}
		log.Debug("BPF map pinning path does not exist. Creating before mounting")
		if err := os.MkdirAll(pinPath, 0700); err != nil {
			return "", fmt.Errorf("creating directory %s: %w", pinPath, err)
		}
	}

	return pinPath, bpfMount(pinPath)
}

func logger() *slog.Logger { return slog.With("component", "ebpf.TracerProvider") }

// this will be just a start node that listens for the context cancellation and then
// unmounts the BPF pinning path
func waitToCloseBbfPinPath(pinPath string) node.StartFuncCtx[[]any] {
	return func(ctx context.Context, _ chan<- []any) {
		<-ctx.Done()
		unmountBpfPinPath(pinPath)
	}
}

func unmountBpfPinPath(pinPath string) {
	log := slog.With("component", "ebpf.TracerProvider", "path", pinPath)
	log.Debug("context has been canceled. Unmounting BPF map pinning path")
	if err := unix.Unmount(pinPath, unix.MNT_FORCE); err != nil {
		log.Warn("can't unmount pinned root. Try unmounting and removing it manually", err)
		return
	}
	log.Debug("unmounted bpf file system")
	if err := os.RemoveAll(pinPath); err != nil {
		log.Warn("can't remove pinned root. Try removing it manually", err)
	} else {
		log.Debug("removed pin path")
	}
}

// filterNotFoundPrograms will filter these programs whose required functions (as
// returned in the Offsets method) haven't been found in the offsets
func filterNotFoundPrograms(programs []Tracer, offsets *goexec.Offsets) []Tracer {
	var filtered []Tracer
	funcs := offsets.Funcs
programs:
	for _, p := range programs {
		for fn, fp := range p.GoProbes() {
			if !fp.Required {
				continue
			}
			if _, ok := funcs[fn]; !ok {
				continue programs
			}
		}
		filtered = append(filtered, p)
	}
	return filtered
}

func allFunctionNames(programs []Tracer) []string {
	uniqueFunctions := map[string]struct{}{}
	var functions []string
	for _, p := range programs {
		for funcName := range p.GoProbes() {
			// avoid duplicating function names
			if _, ok := uniqueFunctions[funcName]; !ok {
				uniqueFunctions[funcName] = struct{}{}
				functions = append(functions, funcName)
			}
		}
	}
	return functions
}

func setGlobalServiceName(ctx context.Context, execElf exec.FileInfo) {
	parts := strings.Split(execElf.CmdExePath, "/")
	global.Context(ctx).ServiceName = parts[len(parts)-1]
}

func inspect(ctx context.Context, cfg *ebpfcommon.TracerConfig, functions []string) (*exec.FileInfo, *goexec.Offsets, error) {
	// Finding the process by port is more complex, it needs to skip proxies written in Go
	if cfg.Port != 0 {
		return inspectPort(ctx, cfg, functions)
	}

	finder := exec.ProcessNamed(cfg.Exec)
	execElf, err := exec.FindExecELF(ctx, finder)
	if err != nil {
		return nil, nil, fmt.Errorf("looking for executable ELF: %w", err)
	}
	defer execElf.ELF.Close()

	var offsets *goexec.Offsets

	if !cfg.SystemWide {
		offsets, err = goexec.InspectOffsets(&execElf, functions)
		if err != nil {
			logger().Info("Go support not detected. Using only generic instrumentation.", "error", err)
		}

		setGlobalServiceName(ctx, execElf)
	}

	return &execElf, offsets, nil
}

// Note: there may be a more efficient way to write this, by potentially passing a validator function to OwnedPort, but
// the code might be a lot harder to follow.
func inspectPort(ctx context.Context, cfg *ebpfcommon.TracerConfig, functions []string) (*exec.FileInfo, *goexec.Offsets, error) {
	invalidPids := make(map[int32]bool)
	for {
		finder := exec.OwnedPort(cfg.Port, invalidPids)

		execElf, err := exec.FindExecELF(ctx, finder)
		if err != nil {
			return nil, nil, fmt.Errorf("looking for executable ELF: %w", err)
		}
		defer execElf.ELF.Close()

		setGlobalServiceName(ctx, execElf)

		offsets, err := goexec.InspectOffsets(&execElf, functions)

		// we didn't find any Go offsets
		// TODO: add an option to keep looking for Go applications if we have a proxy in another language in front of Go
		if err != nil {
			logger().Info("Go support not detected. Using only generic instrumentation.", "error", err)
			return &execElf, offsets, nil
		}

		// we found go offsets, let's see if this application is not a proxy
		for f := range offsets.Funcs {
			// if we find anything of interest other than the Go runtime, we consider this valid application
			if !strings.HasPrefix(f, "runtime.") {
				return &execElf, offsets, nil
			}
		}

		invalidPids[execElf.Pid] = true
	}
}

func printVerifierErrorInfo(err error) {
	var ve *ebpf.VerifierError
	if errors.As(err, &ve) {
		_, _ = fmt.Fprintf(os.Stderr, "Error Log:\n %v\n", strings.Join(ve.Log, "\n"))
	}
}
