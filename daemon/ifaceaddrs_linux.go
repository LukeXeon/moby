//go:build !android

package daemon

import (
	"github.com/moby/moby/v2/daemon/libnetwork/ns"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

// ifaceAddrs returns the addresses from family assigned to the network
// interface with name linkName.
//
// No error is returned if the named interface does not exist.
func ifaceAddrs(linkName string, family int) ([]netlink.Addr, error) {
	nl := ns.NlHandle()
	link, err := nl.LinkByName(linkName)
	if err != nil {
		if !errors.As(err, new(netlink.LinkNotFoundError)) {
			return nil, err
		}
		return nil, nil
	}
	return nl.AddrList(link, family)
}
