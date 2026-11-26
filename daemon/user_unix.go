//go:build !windows && !android

package daemon

import (
	"context"

	coci "github.com/containerd/containerd/v2/pkg/oci"
	"github.com/moby/moby/v2/daemon/container"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func (daemon *Daemon) withUser(c *container.Container) coci.SpecOpts {
	return WithUser(c)
}

func (daemon *Daemon) execUser(ctx context.Context, ec *container.ExecConfig) (specs.User, error) {
	if daemon.UsesSnapshotter() {
		return getUserFromContainerd(ctx, daemon.containerdClient, ec)
	}
	return getUser(ec.Container, ec.User)
}
