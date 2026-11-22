package daemon

import "github.com/vishvananda/netlink"

// ifaceAddrs returns no addresses. On android the bridges of the daemon's
// networks are in the network of the daemon's embedder, which netlink does not
// reach, and an app cannot open netlink sockets.
func ifaceAddrs(string, int) ([]netlink.Addr, error) {
	return nil, nil
}
