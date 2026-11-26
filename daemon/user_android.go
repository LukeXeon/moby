package daemon

import (
	"context"

	"github.com/containerd/containerd/v2/core/containers"
	coci "github.com/containerd/containerd/v2/pkg/oci"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func (daemon *Daemon) withUser(c *container.Container) coci.SpecOpts {
	return func(ctx context.Context, _ coci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Process == nil {
			s.Process = &specs.Process{}
		}
		usr, err := daemon.getUserFromFS(ctx, c, c.Config.User)
		// Preserve fields already set on the process user, such as Umask.
		s.Process.User.UID = usr.UID
		s.Process.User.GID = usr.GID
		s.Process.User.AdditionalGids = usr.AdditionalGids
		return err
	}
}

func (daemon *Daemon) execUser(ctx context.Context, ec *container.ExecConfig) (specs.User, error) {
	return daemon.getUserFromFS(ctx, ec.Container, ec.User)
}

func (daemon *Daemon) lookupContainerUser(ctx context.Context, c *container.Container, username string) (specs.User, error) {
	return daemon.getUserFromFS(ctx, c, username)
}
