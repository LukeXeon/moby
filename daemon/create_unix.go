//go:build !windows

package daemon

import (
	"context"
	"path/filepath"

	"github.com/containerd/log"
	containertypes "github.com/moby/moby/api/types/container"
	mounttypes "github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/moby/v2/daemon/internal/rootids"
	"github.com/moby/moby/v2/daemon/pkg/oci"
	volumeopts "github.com/moby/moby/v2/daemon/volume/service/opts"
	"github.com/opencontainers/selinux/go-selinux/label"
)

// createContainerOSSpecificSettings performs host-OS specific container create functionality
func (daemon *Daemon) createContainerOSSpecificSettings(ctx context.Context, ctr *container.Container) error {
	// Set the default masked and readonly paths with regard to the host config options if they are not set.
	if ctr.HostConfig.MaskedPaths == nil && !ctr.HostConfig.Privileged {
		ctr.HostConfig.MaskedPaths = oci.DefaultSpec().Linux.MaskedPaths // Set it to the default if nil
	}
	if ctr.HostConfig.ReadonlyPaths == nil && !ctr.HostConfig.Privileged {
		ctr.HostConfig.ReadonlyPaths = oci.DefaultSpec().Linux.ReadonlyPaths // Set it to the default if nil
	}
	return nil
}

// createContainerVolumesOS performs host-OS specific volume creation
func (daemon *Daemon) createContainerVolumesOS(ctx context.Context, ctr *container.Container, config *containertypes.Config) error {
	if err := daemon.Mount(ctr); err != nil {
		return err
	}
	defer daemon.Unmount(ctr)

	if err := ctr.SetupWorkingDirectory(rootids.Pair(daemon.idMapping)); err != nil {
		return err
	}

	for spec := range config.Volumes {
		destination := filepath.Clean(spec)

		// Skip volumes for which we already have something mounted on that
		// destination because of a --volume-from.
		if ctr.HasMountFor(destination) {
			log.G(ctx).WithField("container", ctr.ID).WithField("destination", spec).Debug("mountpoint already exists, skipping anonymous volume")
			// Not an error, this could easily have come from the image config.
			continue
		}
		if err := daemon.checkVolumeDestination(ctx, ctr, destination); err != nil {
			return err
		}

		v, err := daemon.volumes.Create(context.TODO(), "", ctr.HostConfig.VolumeDriver, volumeopts.WithCreateReference(ctr.ID))
		if err != nil {
			return err
		}

		if err := label.Relabel(v.Mountpoint, ctr.MountLabel, true); err != nil {
			return err
		}

		ctr.AddMountPointWithVolume(destination, &volumeWrapper{v: v, s: daemon.volumes}, true)
	}
	return daemon.populateVolumes(ctx, ctr)
}

// populateVolumes copies data from the container's rootfs into the volume for non-binds.
// this is only called when the container is created.
func (daemon *Daemon) populateVolumes(ctx context.Context, c *container.Container) error {
	for _, mnt := range c.MountPoints {
		if mnt.Volume == nil {
			continue
		}

		if mnt.Type != mounttypes.TypeVolume || !mnt.CopyData {
			continue
		}

		if err := daemon.populateVolume(ctx, c, mnt); err != nil {
			return err
		}
	}
	return nil
}
