package libnetwork

import (
	"context"
	"fmt"

	"github.com/moby/moby/v2/daemon/libnetwork/config"
	"github.com/moby/moby/v2/daemon/libnetwork/datastore"
	"github.com/moby/moby/v2/daemon/libnetwork/driverapi"
	"github.com/moby/moby/v2/daemon/libnetwork/drivers/host"
	"github.com/moby/moby/v2/daemon/libnetwork/drivers/null"
	"github.com/moby/moby/v2/daemon/libnetwork/drvregistry"
)

// An Android app has no access to netlink, iptables or network namespaces, so
// the Linux bridge, ipvlan, macvlan and overlay drivers and the NAT port mapper
// cannot run. The embedder of the daemon may implement drivers, such as
// bridge, in user space and register them through SetDrivers. The host and
// null drivers only record the network mode.

// RegisterDrivers registers the network drivers of an embedder with the
// controller, with the controller's configuration and store.
type RegisterDrivers func(r driverapi.Registerer, cfg *config.Config, store *datastore.Store) error

var embedderDrivers RegisterDrivers

// SetDrivers sets the network drivers of the embedder. It must be called
// before the daemon starts.
func SetDrivers(register RegisterDrivers) {
	embedderDrivers = register
}

func registerNetworkDrivers(r driverapi.Registerer, cfg *config.Config, store *datastore.Store, _ *drvregistry.PortMappers) error {
	for _, nr := range []struct {
		ntype    string
		register func() error
	}{
		{ntype: host.NetworkType, register: func() error { return host.Register(r) }},
		{ntype: null.NetworkType, register: func() error { return null.Register(r) }},
	} {
		if err := nr.register(); err != nil {
			return fmt.Errorf("failed to register %q driver: %w", nr.ntype, err)
		}
	}
	if embedderDrivers != nil {
		return embedderDrivers(r, cfg, store)
	}
	return nil
}

func registerPortMappers(context.Context, *drvregistry.PortMappers, *config.Config) error {
	return nil
}
