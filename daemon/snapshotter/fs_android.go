package snapshotter

import (
	"context"
	"io"
	"os"

	"github.com/containerd/containerd/v2/core/mount"
	"github.com/moby/go-archive"
	containertypes "github.com/moby/moby/api/types/container"
)

// BindMount describes a host file or directory mounted over a container rootfs.
type BindMount struct {
	Destination string
	Source      string
	ReadOnly    bool
}

// ExtractOptions configures tar extraction into a container directory.
type ExtractOptions struct {
	AllowOverwriteDirWithFile bool
	ChownUID                  *int
	ChownGID                  *int
}

// ContainerFS provides access to a container's merged filesystem on Android,
// where the daemon cannot mount overlayfs or bind mounts on the host.
type ContainerFS interface {
	ReadFile(ctx context.Context, containerID string, running bool, rootfs []mount.Mount, path string) ([]byte, error)
	ExportRootfs(ctx context.Context, containerID string, running bool, rootfs []mount.Mount, w io.Writer) error
	Changes(ctx context.Context, rootfs []mount.Mount) ([]archive.Change, error)
	StatRootfs(ctx context.Context, rootfs []mount.Mount, path string) (os.FileMode, bool, error)
	CopyDir(ctx context.Context, rootfs []mount.Mount, srcPath, destDir string) error
	StatPath(ctx context.Context, containerID string, running bool, rootfs []mount.Mount, mounts []BindMount, path string) (*containertypes.PathStat, error)
	ArchivePath(ctx context.Context, containerID string, running bool, rootfs []mount.Mount, mounts []BindMount, path string, w io.Writer) error
	ExtractToDir(ctx context.Context, containerID string, running bool, rootfs []mount.Mount, mounts []BindMount, readonlyRootfs bool, path string, opts ExtractOptions, r io.Reader) error
}

var containerFS ContainerFS

// SetContainerFS registers the embedder's ContainerFS implementation. It must
// be called before the daemon starts.
func SetContainerFS(fs ContainerFS) {
	containerFS = fs
}

// GetContainerFS returns the registered ContainerFS, or nil if none is set.
func GetContainerFS() ContainerFS {
	return containerFS
}
