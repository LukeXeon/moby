//go:build !no_embedded_containerd

package embedded

import (
	"net"
	"path/filepath"

	"github.com/containerd/plugin"
	"github.com/containerd/ttrpc"
	"github.com/docker/go-connections/sockets"
)

// Android registers no platform-specific containerd plugins; cross-platform
// plugins are registered in server.go.
//
// The program embedding the daemon registers the overlayfs snapshotter, the
// task runtime and the differ: an Android app cannot start shim processes,
// mount filesystems or create device nodes, so containers run in the
// embedder's sandboxes and layers are applied without host mounts. Layer
// directories keep the modes of the image's directories, which an app without
// CAP_DAC_OVERRIDE cannot empty with os.RemoveAll, so the embedder also
// removes them. The cgroups task monitor reads host cgroups and is not
// registered.

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
