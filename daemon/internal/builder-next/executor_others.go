//go:build (!linux && !windows) || android

package buildkit

import (
	"github.com/moby/buildkit/executor"
	"github.com/moby/buildkit/util/network"
)

func newExecutor(executorOpts) (executor.Executor, network.ProxyProvider, error) {
	return &stubExecutor{}, nil, nil
}
