package listeners

import (
	"context"
	"net"
	"os"
	"slices"
	"syscall"

	"github.com/containerd/log"
	"github.com/docker/go-connections/sockets"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// aidShell is the UID of adbd on user builds; "adb forward" connections to an
// abstract socket come from it.
const aidShell = 2000

// listenUnix listens on a socket in the app's private storage, or on an
// abstract socket when addr starts with "@". An app cannot change the owner or
// group of a socket, so socketGroup does not apply.
//
// Every app shares the abstract namespace, so the abstract listener accepts
// only peers running as this app, as root, or as the shell user.
func listenUnix(addr, socketGroup string) (net.Listener, error) {
	if socketGroup != "" && socketGroup != defaultSocketGroup {
		return nil, errors.Errorf("can't set group %s on unix socket %s: not supported on Android", socketGroup, addr)
	}
	if addr == "" || addr[0] != '@' {
		l, err := sockets.NewUnixSocketWithOpts(addr, sockets.WithChmod(0o660))
		if err != nil {
			return nil, errors.Wrapf(err, "can't create unix socket %s", addr)
		}
		return l, nil
	}
	l, err := net.Listen("unix", addr)
	if err != nil {
		return nil, errors.Wrapf(err, "can't create abstract unix socket %s", addr)
	}
	return &peerListener{Listener: l, allowed: []uint32{uint32(os.Getuid()), 0, aidShell}}, nil
}

type peerListener struct {
	net.Listener
	allowed []uint32
}

func (l *peerListener) Accept() (net.Conn, error) {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}
		uid, err := peerUID(conn.(*net.UnixConn))
		if err == nil && slices.Contains(l.allowed, uid) {
			return conn, nil
		}
		log.G(context.TODO()).WithError(err).WithField("uid", uid).Warn("rejected connection to the abstract API socket")
		_ = conn.Close()
	}
}

func peerUID(conn *net.UnixConn) (uint32, error) {
	raw, err := conn.SyscallConn()
	if err != nil {
		return 0, err
	}
	var cred *unix.Ucred
	var credErr error
	if err := raw.Control(func(fd uintptr) {
		cred, credErr = unix.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	}); err != nil {
		return 0, err
	}
	if credErr != nil {
		return 0, credErr
	}
	return cred.Uid, nil
}
