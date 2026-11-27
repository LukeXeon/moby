package snapshotter

import (
	"github.com/containerd/containerd/v2/core/mount"
	"github.com/moby/sys/user"
)

// An Android app cannot mount filesystems. The runtime assembles the root of a
// container from the snapshot mounts in the container's sandbox, and the
// daemon's own access to the container filesystem uses the writable layer
// directory instead of a merged view.
func newMounter(string, string, user.IdentityMapping) Mounter {
	return layerDirMounter{}
}

type layerDirMounter struct{}

func (layerDirMounter) Mount(mounts []mount.Mount, _ string) (string, error) {
	return WritableDir(mounts)
}

func (layerDirMounter) Unmount(string) error { return nil }

func (layerDirMounter) Mounted(string) (string, error) { return "", nil }
