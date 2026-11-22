//go:build !android

package libnetwork

import "net"

// dialInHost connects to a DNS server from the host's network namespace.
func dialInHost(proto, addr string) (net.Conn, error) {
	return net.DialTimeout(proto, addr, extIOTimeout)
}

// dialInSandbox connects to an external DNS server from the network namespace
// of the sandbox.
func (r *Resolver) dialInSandbox(proto, addr string) (net.Conn, error) {
	var (
		extConn net.Conn
		dialErr error
	)
	err := r.backend.ExecFunc(func() {
		extConn, dialErr = net.DialTimeout(proto, addr, extIOTimeout)
	})
	if err != nil {
		return nil, err
	}
	if dialErr != nil {
		return nil, dialErr
	}

	return extConn, nil
}
