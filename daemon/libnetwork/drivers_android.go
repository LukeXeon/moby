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
// cannot run. The host and null drivers only record the network mode.
func registerNetworkDrivers(r driverapi.Registerer, _ *config.Config, _ *datastore.Store, _ *drvregistry.PortMappers) error {
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
	return nil
}

func registerPortMappers(context.Context, *drvregistry.PortMappers, *config.Config) error {
	return nil
}
