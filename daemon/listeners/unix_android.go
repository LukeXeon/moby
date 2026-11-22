package listeners

import (
	"net"

	"github.com/docker/go-connections/sockets"
	"github.com/pkg/errors"
)

// listenUnix listens on a socket in the app's private storage. An app cannot
// change the owner or group of a socket, so socketGroup does not apply.
func listenUnix(addr, socketGroup string) (net.Listener, error) {
	if socketGroup != "" && socketGroup != defaultSocketGroup {
		return nil, errors.Errorf("can't set group %s on unix socket %s: not supported on Android", socketGroup, addr)
	}
	l, err := sockets.NewUnixSocketWithOpts(addr, sockets.WithChmod(0o660))
	if err != nil {
		return nil, errors.Wrapf(err, "can't create unix socket %s", addr)
	}
	return l, nil
}
