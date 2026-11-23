package libnetwork

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/moby/moby/v2/daemon/libnetwork/osl"
)

// On android, the network stack of a task is not the daemon's to enter, so
// the resolver's listeners, which SetupFunc opens through the sandbox's
// InvokeFunc, are in the daemon's network. Instead of DNAT rules in the
// namespace, the sandbox redirects the queries that the task sends to the
// resolver's address to those listeners, and queries to external servers
// leave from the sandbox through its Dial.

func (r *Resolver) setupNAT(context.Context) error {
	if r.err != nil {
		return r.err
	}
	ns, err := r.namespace()
	if err != nil {
		return err
	}
	from := netip.AddrPortFrom(r.listenAddress, 53)
	if err := ns.Redirect("udp", from, r.conn.LocalAddr().(*net.UDPAddr).AddrPort()); err != nil {
		return err
	}
	return ns.Redirect("tcp", from, r.tcpListen.Addr().(*net.TCPAddr).AddrPort())
}

func (r *Resolver) dialInSandbox(proto, addr string) (net.Conn, error) {
	ns, err := r.namespace()
	if err != nil {
		return nil, err
	}
	return ns.Dial(proto, addr)
}

// namespace returns the network namespace of the resolver's sandbox.
func (r *Resolver) namespace() (*osl.Namespace, error) {
	sb, ok := r.backend.(*Sandbox)
	if !ok {
		return nil, fmt.Errorf("resolver backend %T has no network namespace", r.backend)
	}
	sb.mu.Lock()
	ns := sb.osSbox
	sb.mu.Unlock()
	if ns == nil {
		return nil, fmt.Errorf("osl sandbox unavailable for the resolver of %v", sb.ContainerID())
	}
	return ns, nil
}
