//go:build !android

package listeners

import (
	"context"
	"net"
	"os"

	"github.com/containerd/log"
	"github.com/docker/go-connections/sockets"
	"github.com/moby/moby/v2/pkg/homedir"
	"github.com/pkg/errors"
)

func listenUnix(addr, socketGroup string) (net.Listener, error) {
	gid, err := lookupGID(socketGroup)
	if err != nil {
		if socketGroup != "" {
			if socketGroup != defaultSocketGroup {
				return nil, err
			}
			log.G(context.TODO()).Warnf("could not change group %s to %s: %v", addr, defaultSocketGroup, err)
		}
		gid = os.Getgid()
	}
	l, err := sockets.NewUnixSocket(addr, gid)
	if err != nil {
		return nil, errors.Wrapf(err, "can't create unix socket %s", addr)
	}
	if _, err := homedir.StickRuntimeDirContents([]string{addr}); err != nil {
		// StickRuntimeDirContents returns nil error if XDG_RUNTIME_DIR is just unset
		log.G(context.TODO()).WithError(err).Warnf("cannot set sticky bit on socket %s under XDG_RUNTIME_DIR", addr)
	}
	return l, nil
}
