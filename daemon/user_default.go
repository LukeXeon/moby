//go:build !android

package daemon

import (
	"context"

	"github.com/moby/moby/v2/daemon/container"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func (daemon *Daemon) lookupContainerUser(_ context.Context, c *container.Container, username string) (specs.User, error) {
	return getUser(c, username)
}
