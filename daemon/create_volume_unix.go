//go:build !windows && !android

package daemon

import (
	"context"
	"fmt"
	"os"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/log"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/moby/v2/daemon/internal/idtools"
	"github.com/moby/moby/v2/daemon/internal/rootids"
	volumemounts "github.com/moby/moby/v2/daemon/volume/mounts"
	"github.com/pkg/errors"
)

func (daemon *Daemon) checkVolumeDestination(_ context.Context, ctr *container.Container, destination string) error {
	path, err := ctr.GetResourcePath(destination)
	if err != nil {
		return err
	}
	stat, err := os.Stat(path)
	if err == nil && !stat.IsDir() {
		return fmt.Errorf("cannot mount volume over existing file, file exists %s", path)
	}
	return nil
}

func (daemon *Daemon) populateVolume(ctx context.Context, c *container.Container, mnt *volumemounts.MountPoint) error {
	ctrDestPath, err := c.GetResourcePath(mnt.Destination)
	if err != nil {
		return err
	}

	stat, err := os.Stat(ctrDestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Volume population copies directory contents from the container image
	// into the volume. This only applies when the source is a directory.
	if !stat.IsDir() {
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
	if err := c.CopyImagePathContent(volumePath, ctrDestPath); err != nil {
		return err
	}

	return nil
}
