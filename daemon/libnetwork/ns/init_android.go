package ns

import (
	"github.com/moby/moby/v2/daemon/libnetwork/nlwrap"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// An Android app cannot open netlink sockets, and the network namespaces of
// the daemon's containers are not kernel namespaces, see osl. The handle
// opens a socket for each request, as the package-level functions of netlink
// do, so a request fails with the kernel's error instead of the daemon
// panicking when it first needs a handle.

// ResetHandles does nothing: the handles keep no state.
func ResetHandles() {}

// NsHandle returns the network namespace handle for the initial (host) namespace.
func NsHandle() netns.NsHandle {
	return netns.None()
}

// NlHandle returns the netlink handle.
func NlHandle() nlwrap.Handle {
	return nlwrap.Handle{Handle: &netlink.Handle{}}
}
