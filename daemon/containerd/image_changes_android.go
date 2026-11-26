package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/errdefs"
	"github.com/moby/go-archive"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/moby/v2/daemon/snapshotter"
)

func (i *ImageService) Changes(ctx context.Context, ctr *container.Container) ([]archive.Change, error) {
	fs := snapshotter.GetContainerFS()
	if fs == nil {
		return nil, fmt.Errorf("container filesystem is not configured: %w", errdefs.ErrNotImplemented)
	}
	if ctr.RWLayer == nil {
		return nil, fmt.Errorf("RWLayer is unexpectedly nil for container %s", ctr.ID)
	}
	sn := i.client.SnapshotService(ctr.Driver)
	mounts, err := sn.Mounts(ctx, ctr.ID)
	if err != nil {
		return nil, err
	}
	return fs.Changes(ctx, mounts)
}
