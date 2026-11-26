package daemon

import (
	"context"
	"io"

	"github.com/moby/moby/api/types/events"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/moby/v2/errdefs"
)

func (daemon *Daemon) containerExport(ctx context.Context, ctr *container.Container, out io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fs, err := daemon.containerFS()
	if err != nil {
		return err
	}
	rootfs, err := daemon.containerRootfsMounts(ctx, ctr)
	if err != nil {
		return err
	}
	if err := fs.ExportRootfs(ctx, ctr.ID, ctr.State.Running, rootfs, out); err != nil {
		if ctx.Err() != nil {
			return errdefs.Cancelled(ctx.Err())
		}
		return err
	}
	daemon.LogContainerEvent(ctr, events.ActionExport)
	return nil
}
