package daemon

import (
	"context"
	"fmt"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/log"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/moby/v2/daemon/internal/idtools"
	"github.com/moby/moby/v2/daemon/internal/rootids"
	volumemounts "github.com/moby/moby/v2/daemon/volume/mounts"
	"github.com/pkg/errors"
)

func (daemon *Daemon) checkVolumeDestination(ctx context.Context, ctr *container.Container, destination string) error {
	fs, err := daemon.containerFS()
	if err != nil {
		return err
	}
	rootfs, err := daemon.containerRootfsMounts(ctx, ctr)
	if err != nil {
		return err
	}
	mode, exists, err := fs.StatRootfs(ctx, rootfs, destination)
	if err != nil {
		return err
	}
	if exists && !mode.IsDir() {
		return fmt.Errorf("cannot mount volume over existing file, file exists %s", destination)
	}
	return nil
}

func (daemon *Daemon) populateVolume(ctx context.Context, c *container.Container, mnt *volumemounts.MountPoint) error {
	fs, err := daemon.containerFS()
	if err != nil {
		return err
	}
	rootfs, err := daemon.containerRootfsMounts(ctx, c)
	if err != nil {
		return err
	}
	mode, exists, err := fs.StatRootfs(ctx, rootfs, mnt.Destination)
	if err != nil {
		return err
	}
	if !exists || !mode.IsDir() {
		return nil
	}

	uid, gid := rootids.Pair(daemon.idMapping)
	volumePath, cleanup, err := mnt.Setup(ctx, c.MountLabel, idtools.Identity{UID: uid, GID: gid}, nil)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return nil
		}
		log.G(ctx).WithError(err).Debugf("can't copy data from %s:%s, to %s", c.ID, mnt.Destination, volumePath)
		return errors.Wrapf(err, "failed to populate volume")
	}
	defer func() {
		ctx := context.WithoutCancel(ctx)
		_ = cleanup(ctx)
		_ = mnt.Cleanup(ctx)
	}()

	log.G(ctx).Debugf("copying image data from %s:%s, to %s", c.ID, mnt.Destination, volumePath)
	return fs.CopyDir(ctx, rootfs, mnt.Destination, volumePath)
}
