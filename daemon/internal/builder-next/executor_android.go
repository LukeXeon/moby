package buildkit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/containerd/log"
	"github.com/moby/buildkit/executor"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/util/network"
	"github.com/moby/moby/v2/daemon/builder"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const networkName = "bridge"

func newExecutor(opts executorOpts) (executor.Executor, network.ProxyProvider, error) {
	factory := builder.GetBuildkitExecutorFactory()
	if factory == nil {
		return nil, nil, errors.New("buildkit executor factory is not registered on android")
	}

	netRoot := filepath.Join(opts.root, "net")
	if fis, err := os.ReadDir(netRoot); err == nil {
		for _, fi := range fis {
			fp := filepath.Join(netRoot, fi.Name())
			if err := os.RemoveAll(fp); err != nil {
				log.G(context.TODO()).WithError(err).Errorf("failed to delete old network state: %v", fp)
			}
		}
	}

	networkProviders := map[pb.NetMode]network.Provider{
		pb.NetMode_UNSET: &bridgeProvider{Controller: opts.networkController, Hooks: opts.ociHooks, Root: netRoot},
		pb.NetMode_HOST:  network.NewHostProvider(),
		pb.NetMode_NONE:  network.NewNoneProvider(),
	}

	exec, err := factory(builder.BuildkitExecutorConfig{
		Root:                filepath.Join(opts.root, "executor"),
		NetworkProviders:    networkProviders,
		DNSConfig:           opts.dnsConfig,
		ContainerdAddr:      opts.containerdAddr,
		ContainerdDialer:    opts.containerdDialer,
		ContainerdNamespace: opts.containerdNamespace,
	})
	if err != nil {
		return nil, nil, err
	}
	return exec, nil, nil
}

// Set waits for the libnetwork sandbox to be ready. On Android there are no
// OCI prestart hook binaries; the executor attaches the sandbox in
// OnCreateRuntime once the Sentry task has a PID.
func (iface *lnInterface) Set(*specs.Spec) error {
	<-iface.ready
	return iface.err
}

// OnCreateRuntime attaches the libnetwork sandbox to the newly created Sentry
// task identified by pid.
func (iface *lnInterface) OnCreateRuntime(pid uint32) error {
	<-iface.ready
	if iface.err != nil {
		return iface.err
	}
	return iface.sbx.SetKey(iface.ctx, fmt.Sprintf("/proc/%d/ns/net", pid))
}
