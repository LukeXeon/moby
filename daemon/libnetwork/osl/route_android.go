package osl

import (
	"net"
	"slices"

	"github.com/moby/moby/v2/daemon/libnetwork/types"
)

// SetGateway sets the default IPv4 gateway for the sandbox. It is a no-op
// if the given gateway is empty.
func (n *Namespace) SetGateway(gw net.IP) error {
	if len(gw) == 0 {
		return nil
	}
	return n.notImplemented("setting gateway " + gw.String())
}

// UnsetGateway the previously set default IPv4 gateway in the sandbox.
func (n *Namespace) UnsetGateway() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.gw = net.IP{}
	return nil
}

// SetGatewayIPv6 sets the default IPv6 gateway for the sandbox. It is a no-op
// if the given gateway is empty.
func (n *Namespace) SetGatewayIPv6(gwv6 net.IP) error {
	if len(gwv6) == 0 {
		return nil
	}
	return n.notImplemented("setting gateway " + gwv6.String())
}

// UnsetGatewayIPv6 unsets the previously set default IPv6 gateway in the sandbox.
func (n *Namespace) UnsetGatewayIPv6() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.gwv6 = net.IP{}
	return nil
}

// AddStaticRoute adds a static route to the sandbox.
func (n *Namespace) AddStaticRoute(r *types.StaticRoute) error {
	return n.notImplemented("adding route " + r.Destination.String())
}

// RemoveStaticRoute removes a static route from the sandbox.
func (n *Namespace) RemoveStaticRoute(r *types.StaticRoute) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.staticRoutes = slices.DeleteFunc(n.staticRoutes, func(route *types.StaticRoute) bool {
		return route == r
	})
	return nil
}

// SetDefaultRouteIPv4 sets up a connected route to 0.0.0.0 via the Interface
// with srcName.
func (n *Namespace) SetDefaultRouteIPv4(srcName string) error {
	return n.notImplemented("setting the IPv4 default route to " + srcName)
}

// SetDefaultRouteIPv6 sets up a connected route to [::] via the Interface
// with srcName.
func (n *Namespace) SetDefaultRouteIPv6(srcName string) error {
	return n.notImplemented("setting the IPv6 default route to " + srcName)
}

// UnsetDefaultRouteIPv4 unsets the previously set default IPv4 default route
// in the sandbox.
func (n *Namespace) UnsetDefaultRouteIPv4() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.defRoute4SrcName = ""
	return nil
}

// UnsetDefaultRouteIPv6 unsets the previously set default IPv6 default route
// in the sandbox.
func (n *Namespace) UnsetDefaultRouteIPv6() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.defRoute6SrcName = ""
	return nil
}
