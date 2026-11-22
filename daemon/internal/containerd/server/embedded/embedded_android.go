//go:build !no_embedded_containerd

package embedded

import (
	"net"
	"path/filepath"

	"github.com/containerd/plugin"
	"github.com/containerd/ttrpc"
	"github.com/docker/go-connections/sockets"

	// Android-specific containerd plugin registrations: the overlayfs
	// snapshotter, which only manages layer directories here.
	//
	// The program embedding the daemon registers the task runtime and the
	// differ: an Android app cannot start shim processes, mount filesystems or
	// create device nodes, so containers run in the embedder's sandboxes and
	// layers are applied without host mounts. The cgroups task monitor reads
	// host cgroups and is not registered.
	//
	// Cross-platform plugins are registered in server.go.
	_ "github.com/containerd/containerd/v2/plugins/snapshots/overlay/plugin"
)

// runsShims reports whether tasks run in shim processes, whose sockets and
// FIFOs use the default state directory.
const runsShims = false

func defaultAddress(stateDir string) string {
	return filepath.Join(stateDir, "containerd.sock")
}

func listen(address string) (net.Listener, error) {
	return sockets.NewUnixSocketWithOpts(address, sockets.WithChmod(0o660))
}

func newTTRPCServer() (*ttrpc.Server, error) {
	return ttrpc.NewServer(
		ttrpc.WithServerHandshaker(ttrpc.UnixSocketRequireSameUser()),
	)
}

func disableCgroupsPrometheus(registration plugin.Registration) any {
	return registration.Config
}
