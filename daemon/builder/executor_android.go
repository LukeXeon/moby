//go:build android

package builder

import (
	"context"
	"net"
	"sync"

	"github.com/moby/buildkit/executor"
	"github.com/moby/buildkit/executor/oci"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/util/network"
)

// BuildkitExecutorConfig supplies the dependencies needed by the embedder's
// Android BuildKit executor.
type BuildkitExecutorConfig struct {
	Root                string
	NetworkProviders    map[pb.NetMode]network.Provider
	DNSConfig           *oci.DNSConfig
	ContainerdAddr      string
	ContainerdDialer    func(ctx context.Context, address string) (net.Conn, error)
	ContainerdNamespace string
}

// BuildkitExecutorFactory builds the BuildKit executor on Android.
type BuildkitExecutorFactory func(BuildkitExecutorConfig) (executor.Executor, error)

var (
	buildkitExecutorMu      sync.RWMutex
	buildkitExecutorFactory BuildkitExecutorFactory
)

// SetBuildkitExecutorFactory registers the Android BuildKit executor
// constructor.
func SetBuildkitExecutorFactory(fn BuildkitExecutorFactory) {
	buildkitExecutorMu.Lock()
	buildkitExecutorFactory = fn
	buildkitExecutorMu.Unlock()
}

// GetBuildkitExecutorFactory returns the registered Android BuildKit executor
// constructor, if any.
func GetBuildkitExecutorFactory() BuildkitExecutorFactory {
	buildkitExecutorMu.RLock()
	defer buildkitExecutorMu.RUnlock()
	return buildkitExecutorFactory
}
