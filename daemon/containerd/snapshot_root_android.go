package containerd

import (
	"context"

	"github.com/containerd/containerd/v2/core/mount"
	"github.com/moby/moby/v2/daemon/snapshotter"
)

// withSnapshotRoot calls fn with the writable layer directory of a snapshot:
// an Android app cannot mount it. fn sees the files of this layer only, not
// those of its parents.
func withSnapshotRoot(_ context.Context, mounts []mount.Mount, fn func(root string) error) error {
	dir, err := snapshotter.WritableDir(mounts)
	if err != nil {
		return err
	}
	return fn(dir)
}
