package daemon

import (
	"context"
	"io"
	"path/filepath"

	"github.com/moby/go-archive"
	"github.com/moby/go-archive/compression"
	containertypes "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/events"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/moby/moby/v2/daemon/snapshotter"
	"github.com/moby/moby/v2/errdefs"
	"github.com/moby/moby/v2/pkg/ioutils"
)

func (daemon *Daemon) containerStatPath(ctr *container.Container, path string) (*containertypes.PathStat, error) {
	ctr.Lock()
	defer ctr.Unlock()

	ctx := context.Background()
	fs, err := daemon.containerFS()
	if err != nil {
		return nil, err
	}
	rootfs, err := daemon.containerRootfsMounts(ctx, ctr)
	if err != nil {
		return nil, err
	}
	mounts, cleanup, err := daemon.containerBindMounts(ctx, ctr)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	absPath := archive.PreserveTrailingDotOrSeparator(filepath.Join("/", path), path)
	return fs.StatPath(ctx, ctr.ID, ctr.State.Running, rootfs, mounts, absPath)
}

func (daemon *Daemon) containerArchivePath(ctr *container.Container, path string) (content io.ReadCloser, stat *containertypes.PathStat, retErr error) {
	ctr.Lock()
	defer func() {
		if retErr != nil {
			ctr.Unlock()
		}
	}()

	ctx := context.Background()
	fs, err := daemon.containerFS()
	if err != nil {
		return nil, nil, err
	}
	rootfs, err := daemon.containerRootfsMounts(ctx, ctr)
	if err != nil {
		return nil, nil, err
	}
	mounts, cleanup, err := daemon.containerBindMounts(ctx, ctr)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if retErr != nil {
			cleanup()
		}
	}()

	absPath := archive.PreserveTrailingDotOrSeparator(filepath.Join("/", path), path)
	running := ctr.State.Running
	stat, err = fs.StatPath(ctx, ctr.ID, running, rootfs, mounts, absPath)
	if err != nil {
		return nil, nil, err
	}

	pr, pw := io.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = pw.CloseWithError(fs.ArchivePath(ctx, ctr.ID, running, rootfs, mounts, absPath, pw))
	}()

	content = ioutils.NewReadCloserWrapper(pr, func() error {
		err := pr.Close()
		<-done
		cleanup()
		ctr.Unlock()
		return err
	})

	daemon.LogContainerEvent(ctr, events.ActionArchivePath)
	return content, stat, nil
}

func (daemon *Daemon) containerExtractToDir(ctr *container.Container, path string, copyUIDGID, allowOverwriteDirWithFile bool, content io.Reader) error {
	ctr.Lock()
	defer ctr.Unlock()

	ctx := context.Background()
	fs, err := daemon.containerFS()
	if err != nil {
		return err
	}
	rootfs, err := daemon.containerRootfsMounts(ctx, ctr)
	if err != nil {
		return err
	}
	mounts, cleanup, err := daemon.containerBindMounts(ctx, ctr)
	if err != nil {
		return err
	}
	defer cleanup()

	decompressed, err := compression.DecompressStream(content)
	if err != nil {
		return err
	}
	defer decompressed.Close()

	opts := snapshotter.ExtractOptions{
		AllowOverwriteDirWithFile: allowOverwriteDirWithFile,
	}
	if copyUIDGID && ctr.Config.User != "" {
		usr, err := daemon.getUserFromFS(ctx, ctr, ctr.Config.User)
		if err != nil {
			return errdefs.InvalidParameter(err)
		}
		uid, gid := int(usr.UID), int(usr.GID)
		opts.ChownUID = &uid
		opts.ChownGID = &gid
	}

	absPath := archive.PreserveTrailingDotOrSeparator(filepath.Join("/", path), path)
	if err := fs.ExtractToDir(ctx, ctr.ID, ctr.State.Running, rootfs, mounts, ctr.HostConfig.ReadonlyRootfs, absPath, opts, decompressed); err != nil {
		return err
	}

	daemon.LogContainerEvent(ctr, events.ActionExtractToDir)
	return nil
}
