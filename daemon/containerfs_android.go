package daemon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/containerd/containerd/v2/core/mount"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/containerd/log"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/moby/v2/daemon/snapshotter"
	"github.com/moby/sys/user"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type rootfsMountProvider interface {
	Mounts(context.Context) ([]mount.Mount, error)
}

func (daemon *Daemon) containerFS() (snapshotter.ContainerFS, error) {
	fs := snapshotter.GetContainerFS()
	if fs == nil {
		return nil, fmt.Errorf("container filesystem is not configured: %w", cerrdefs.ErrNotImplemented)
	}
	return fs, nil
}

func (daemon *Daemon) containerRootfsMounts(ctx context.Context, ctr *container.Container) ([]mount.Mount, error) {
	if ctr.RWLayer == nil {
		return nil, fmt.Errorf("container %s has no rootfs", ctr.ID)
	}
	p, ok := ctr.RWLayer.(rootfsMountProvider)
	if !ok {
		return nil, fmt.Errorf("container %s layer %T does not expose mounts", ctr.ID, ctr.RWLayer)
	}
	return p.Mounts(ctx)
}

func (daemon *Daemon) containerBindMounts(ctx context.Context, ctr *container.Container) ([]snapshotter.BindMount, func(), error) {
	rawMounts, cleanupMounts, err := daemon.setupMounts(ctx, ctr)
	if err != nil {
		return nil, nil, err
	}
	mounts := make([]snapshotter.BindMount, 0, len(rawMounts))
	for _, m := range rawMounts {
		mounts = append(mounts, snapshotter.BindMount{
			Destination: m.Destination,
			Source:      m.Source,
			ReadOnly:    !m.Writable,
		})
	}
	cleanup := func() {
		bg := context.WithoutCancel(ctx)
		if err := cleanupMounts(bg); err != nil {
			log.G(bg).WithError(err).Debug("Failed to cleanup container mounts")
		}
		if err := ctr.UnmountVolumes(bg, daemon.LogVolumeEvent); err != nil {
			log.G(bg).WithError(err).Debug("Failed to unmount container volumes")
		}
	}
	return mounts, cleanup, nil
}

func (daemon *Daemon) readContainerFileOptional(ctx context.Context, ctr *container.Container, rootfs []mount.Mount, path string) (io.Reader, error) {
	fs, err := daemon.containerFS()
	if err != nil {
		return nil, err
	}
	data, err := fs.ReadFile(ctx, ctr.ID, ctr.State.Running, rootfs, path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (daemon *Daemon) getUserFromFS(ctx context.Context, ctr *container.Container, username string) (specs.User, error) {
	var usr specs.User
	rootfs, err := daemon.containerRootfsMounts(ctx, ctr)
	if err != nil {
		return usr, err
	}
	passwdPath, err := user.GetPasswdPath()
	if err != nil {
		return usr, err
	}
	groupPath, err := user.GetGroupPath()
	if err != nil {
		return usr, err
	}
	passwdReader, err := daemon.readContainerFileOptional(ctx, ctr, rootfs, passwdPath)
	if err != nil {
		return usr, err
	}
	groupBytes, err := func() ([]byte, error) {
		fs, err := daemon.containerFS()
		if err != nil {
			return nil, err
		}
		data, err := fs.ReadFile(ctx, ctr.ID, ctr.State.Running, rootfs, groupPath)
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return data, err
	}()
	if err != nil {
		return usr, err
	}
	var groupReader io.Reader
	if groupBytes != nil {
		groupReader = bytes.NewReader(groupBytes)
	}
	execUser, err := user.GetExecUser(username, nil, passwdReader, groupReader)
	if err != nil {
		return usr, err
	}
	usr.UID = uint32(execUser.Uid)
	usr.GID = uint32(execUser.Gid)
	usr.AdditionalGids = []uint32{usr.GID}

	var addGroups []int
	if len(ctr.HostConfig.GroupAdd) > 0 {
		var addGroupReader io.Reader
		if groupBytes != nil {
			addGroupReader = bytes.NewReader(groupBytes)
		}
		addGroups, err = user.GetAdditionalGroups(ctr.HostConfig.GroupAdd, addGroupReader)
		if err != nil {
			return usr, err
		}
	}
	for _, g := range append(execUser.Sgids, addGroups...) {
		usr.AdditionalGids = append(usr.AdditionalGids, uint32(g))
	}
	return usr, nil
}
