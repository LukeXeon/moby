//go:build !android

package containerd

import (
	"context"

	"github.com/containerd/containerd/v2/core/mount"
)

// withSnapshotRoot calls fn with the root of a snapshot mounted at a
// temporary directory.
func withSnapshotRoot(ctx context.Context, mounts []mount.Mount, fn func(root string) error) error {
	return mount.WithTempMount(ctx, mounts, fn)
}
