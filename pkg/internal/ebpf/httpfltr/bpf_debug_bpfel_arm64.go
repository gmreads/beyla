// Code generated by bpf2go; DO NOT EDIT.
//go:build arm64
// +build arm64

package httpfltr

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/cilium/ebpf"
)

type bpf_debugConnectionInfoT struct {
	S_addr [16]uint8
	D_addr [16]uint8
	S_port uint16
	D_port uint16
}

type bpf_debugHttpBufT struct {
	Flags    uint64
	ConnInfo bpf_debugConnectionInfoT
	Buf      [1024]uint8
	_        [4]byte
}

type bpf_debugHttpConnectionMetadataT struct {
	Pid struct {
		HostPid   uint32
		UserPid   uint32
		Namespace uint32
	}
	Type uint8
}

type bpf_debugHttpInfoT struct {
	Flags           uint64
	ConnInfo        bpf_debugConnectionInfoT
	_               [4]byte
	StartMonotimeNs uint64
	EndMonotimeNs   uint64
	Buf             [160]uint8
	Len             uint32
	RespLen         uint32
	Status          uint16
	Type            uint8
	Ssl             uint8
	Pid             struct {
		HostPid   uint32
		UserPid   uint32
		Namespace uint32
	}
}

type bpf_debugRecvArgsT struct {
	SockPtr  uint64
	IovecPtr uint64
}

type bpf_debugSockArgsT struct {
	Addr       uint64
	AcceptTime uint64
}

type bpf_debugSslArgsT struct {
	Ssl    uint64
	Buf    uint64
	LenPtr uint64
}

// loadBpf_debug returns the embedded CollectionSpec for bpf_debug.
func loadBpf_debug() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_Bpf_debugBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load bpf_debug: %w", err)
	}

	return spec, err
}

// loadBpf_debugObjects loads bpf_debug and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//	*bpf_debugObjects
//	*bpf_debugPrograms
//	*bpf_debugMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func loadBpf_debugObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := loadBpf_debug()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// bpf_debugSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpf_debugSpecs struct {
	bpf_debugProgramSpecs
	bpf_debugMapSpecs
}

// bpf_debugSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpf_debugProgramSpecs struct {
	KprobeSysExit           *ebpf.ProgramSpec `ebpf:"kprobe_sys_exit"`
	KprobeTcpConnect        *ebpf.ProgramSpec `ebpf:"kprobe_tcp_connect"`
	KprobeTcpRcvEstablished *ebpf.ProgramSpec `ebpf:"kprobe_tcp_rcv_established"`
	KprobeTcpRecvmsg        *ebpf.ProgramSpec `ebpf:"kprobe_tcp_recvmsg"`
	KprobeTcpSendmsg        *ebpf.ProgramSpec `ebpf:"kprobe_tcp_sendmsg"`
	KretprobeSockAlloc      *ebpf.ProgramSpec `ebpf:"kretprobe_sock_alloc"`
	KretprobeSysAccept4     *ebpf.ProgramSpec `ebpf:"kretprobe_sys_accept4"`
	KretprobeSysConnect     *ebpf.ProgramSpec `ebpf:"kretprobe_sys_connect"`
	KretprobeTcpRecvmsg     *ebpf.ProgramSpec `ebpf:"kretprobe_tcp_recvmsg"`
	UprobeSslDoHandshake    *ebpf.ProgramSpec `ebpf:"uprobe_ssl_do_handshake"`
	UprobeSslRead           *ebpf.ProgramSpec `ebpf:"uprobe_ssl_read"`
	UprobeSslReadEx         *ebpf.ProgramSpec `ebpf:"uprobe_ssl_read_ex"`
	UprobeSslShutdown       *ebpf.ProgramSpec `ebpf:"uprobe_ssl_shutdown"`
	UprobeSslWrite          *ebpf.ProgramSpec `ebpf:"uprobe_ssl_write"`
	UprobeSslWriteEx        *ebpf.ProgramSpec `ebpf:"uprobe_ssl_write_ex"`
	UretprobeSslDoHandshake *ebpf.ProgramSpec `ebpf:"uretprobe_ssl_do_handshake"`
	UretprobeSslRead        *ebpf.ProgramSpec `ebpf:"uretprobe_ssl_read"`
	UretprobeSslReadEx      *ebpf.ProgramSpec `ebpf:"uretprobe_ssl_read_ex"`
	UretprobeSslWrite       *ebpf.ProgramSpec `ebpf:"uretprobe_ssl_write"`
	UretprobeSslWriteEx     *ebpf.ProgramSpec `ebpf:"uretprobe_ssl_write_ex"`
}

// bpf_debugMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type bpf_debugMapSpecs struct {
	ActiveAcceptArgs    *ebpf.MapSpec `ebpf:"active_accept_args"`
	ActiveConnectArgs   *ebpf.MapSpec `ebpf:"active_connect_args"`
	ActiveRecvArgs      *ebpf.MapSpec `ebpf:"active_recv_args"`
	ActiveSslHandshakes *ebpf.MapSpec `ebpf:"active_ssl_handshakes"`
	ActiveSslReadArgs   *ebpf.MapSpec `ebpf:"active_ssl_read_args"`
	ActiveSslWriteArgs  *ebpf.MapSpec `ebpf:"active_ssl_write_args"`
	DeadPids            *ebpf.MapSpec `ebpf:"dead_pids"`
	Events              *ebpf.MapSpec `ebpf:"events"`
	FilteredConnections *ebpf.MapSpec `ebpf:"filtered_connections"`
	HttpInfoMem         *ebpf.MapSpec `ebpf:"http_info_mem"`
	OngoingHttp         *ebpf.MapSpec `ebpf:"ongoing_http"`
	PidCache            *ebpf.MapSpec `ebpf:"pid_cache"`
	PidTidToConn        *ebpf.MapSpec `ebpf:"pid_tid_to_conn"`
	SslToConn           *ebpf.MapSpec `ebpf:"ssl_to_conn"`
	SslToPidTid         *ebpf.MapSpec `ebpf:"ssl_to_pid_tid"`
	ValidPids           *ebpf.MapSpec `ebpf:"valid_pids"`
}

// bpf_debugObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to loadBpf_debugObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpf_debugObjects struct {
	bpf_debugPrograms
	bpf_debugMaps
}

func (o *bpf_debugObjects) Close() error {
	return _Bpf_debugClose(
		&o.bpf_debugPrograms,
		&o.bpf_debugMaps,
	)
}

// bpf_debugMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to loadBpf_debugObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpf_debugMaps struct {
	ActiveAcceptArgs    *ebpf.Map `ebpf:"active_accept_args"`
	ActiveConnectArgs   *ebpf.Map `ebpf:"active_connect_args"`
	ActiveRecvArgs      *ebpf.Map `ebpf:"active_recv_args"`
	ActiveSslHandshakes *ebpf.Map `ebpf:"active_ssl_handshakes"`
	ActiveSslReadArgs   *ebpf.Map `ebpf:"active_ssl_read_args"`
	ActiveSslWriteArgs  *ebpf.Map `ebpf:"active_ssl_write_args"`
	DeadPids            *ebpf.Map `ebpf:"dead_pids"`
	Events              *ebpf.Map `ebpf:"events"`
	FilteredConnections *ebpf.Map `ebpf:"filtered_connections"`
	HttpInfoMem         *ebpf.Map `ebpf:"http_info_mem"`
	OngoingHttp         *ebpf.Map `ebpf:"ongoing_http"`
	PidCache            *ebpf.Map `ebpf:"pid_cache"`
	PidTidToConn        *ebpf.Map `ebpf:"pid_tid_to_conn"`
	SslToConn           *ebpf.Map `ebpf:"ssl_to_conn"`
	SslToPidTid         *ebpf.Map `ebpf:"ssl_to_pid_tid"`
	ValidPids           *ebpf.Map `ebpf:"valid_pids"`
}

func (m *bpf_debugMaps) Close() error {
	return _Bpf_debugClose(
		m.ActiveAcceptArgs,
		m.ActiveConnectArgs,
		m.ActiveRecvArgs,
		m.ActiveSslHandshakes,
		m.ActiveSslReadArgs,
		m.ActiveSslWriteArgs,
		m.DeadPids,
		m.Events,
		m.FilteredConnections,
		m.HttpInfoMem,
		m.OngoingHttp,
		m.PidCache,
		m.PidTidToConn,
		m.SslToConn,
		m.SslToPidTid,
		m.ValidPids,
	)
}

// bpf_debugPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadBpf_debugObjects or ebpf.CollectionSpec.LoadAndAssign.
type bpf_debugPrograms struct {
	KprobeSysExit           *ebpf.Program `ebpf:"kprobe_sys_exit"`
	KprobeTcpConnect        *ebpf.Program `ebpf:"kprobe_tcp_connect"`
	KprobeTcpRcvEstablished *ebpf.Program `ebpf:"kprobe_tcp_rcv_established"`
	KprobeTcpRecvmsg        *ebpf.Program `ebpf:"kprobe_tcp_recvmsg"`
	KprobeTcpSendmsg        *ebpf.Program `ebpf:"kprobe_tcp_sendmsg"`
	KretprobeSockAlloc      *ebpf.Program `ebpf:"kretprobe_sock_alloc"`
	KretprobeSysAccept4     *ebpf.Program `ebpf:"kretprobe_sys_accept4"`
	KretprobeSysConnect     *ebpf.Program `ebpf:"kretprobe_sys_connect"`
	KretprobeTcpRecvmsg     *ebpf.Program `ebpf:"kretprobe_tcp_recvmsg"`
	UprobeSslDoHandshake    *ebpf.Program `ebpf:"uprobe_ssl_do_handshake"`
	UprobeSslRead           *ebpf.Program `ebpf:"uprobe_ssl_read"`
	UprobeSslReadEx         *ebpf.Program `ebpf:"uprobe_ssl_read_ex"`
	UprobeSslShutdown       *ebpf.Program `ebpf:"uprobe_ssl_shutdown"`
	UprobeSslWrite          *ebpf.Program `ebpf:"uprobe_ssl_write"`
	UprobeSslWriteEx        *ebpf.Program `ebpf:"uprobe_ssl_write_ex"`
	UretprobeSslDoHandshake *ebpf.Program `ebpf:"uretprobe_ssl_do_handshake"`
	UretprobeSslRead        *ebpf.Program `ebpf:"uretprobe_ssl_read"`
	UretprobeSslReadEx      *ebpf.Program `ebpf:"uretprobe_ssl_read_ex"`
	UretprobeSslWrite       *ebpf.Program `ebpf:"uretprobe_ssl_write"`
	UretprobeSslWriteEx     *ebpf.Program `ebpf:"uretprobe_ssl_write_ex"`
}

func (p *bpf_debugPrograms) Close() error {
	return _Bpf_debugClose(
		p.KprobeSysExit,
		p.KprobeTcpConnect,
		p.KprobeTcpRcvEstablished,
		p.KprobeTcpRecvmsg,
		p.KprobeTcpSendmsg,
		p.KretprobeSockAlloc,
		p.KretprobeSysAccept4,
		p.KretprobeSysConnect,
		p.KretprobeTcpRecvmsg,
		p.UprobeSslDoHandshake,
		p.UprobeSslRead,
		p.UprobeSslReadEx,
		p.UprobeSslShutdown,
		p.UprobeSslWrite,
		p.UprobeSslWriteEx,
		p.UretprobeSslDoHandshake,
		p.UretprobeSslRead,
		p.UretprobeSslReadEx,
		p.UretprobeSslWrite,
		p.UretprobeSslWriteEx,
	)
}

func _Bpf_debugClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//
//go:embed bpf_debug_bpfel_arm64.o
var _Bpf_debugBytes []byte
