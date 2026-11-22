package osl

import (
	"fmt"
	"net"
	"slices"

	"github.com/moby/moby/v2/daemon/libnetwork/types"
)

var (
	defaultRouteIPv4 = &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}
	defaultRouteIPv6 = &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}
)

// SetGateway sets the default IPv4 gateway for the sandbox. It is a no-op
// if the given gateway is empty.
func (n *Namespace) SetGateway(gw net.IP) error {
	if len(gw) == 0 {
		return nil
	}
	if err := n.programRoute(defaultRouteIPv4, gw, true); err != nil {
		return err
	}
	n.mu.Lock()
	n.gw = gw
	n.mu.Unlock()
	return nil
}

// UnsetGateway the previously set default IPv4 gateway in the sandbox.
// It is a no-op if no gateway was set.
func (n *Namespace) UnsetGateway() error {
	n.mu.Lock()
	gw := n.gw
	n.mu.Unlock()
	if len(gw) == 0 {
		return nil
	}
	if err := n.programRoute(defaultRouteIPv4, gw, false); err != nil {
		return err
	}
	n.mu.Lock()
	n.gw = net.IP{}
	n.mu.Unlock()
	return nil
}

// SetGatewayIPv6 sets the default IPv6 gateway for the sandbox. It is a no-op
// if the given gateway is empty.
func (n *Namespace) SetGatewayIPv6(gwv6 net.IP) error {
	if len(gwv6) == 0 {
		return nil
	}
	if err := n.programRoute(defaultRouteIPv6, gwv6, true); err != nil {
		return err
	}
	n.mu.Lock()
	n.gwv6 = gwv6
	n.mu.Unlock()
	return nil
}

// UnsetGatewayIPv6 unsets the previously set default IPv6 gateway in the sandbox.
// It is a no-op if no gateway was set.
func (n *Namespace) UnsetGatewayIPv6() error {
	n.mu.Lock()
	gwv6 := n.gwv6
	n.mu.Unlock()
	if len(gwv6) == 0 {
		return nil
	}
	if err := n.programRoute(defaultRouteIPv6, gwv6, false); err != nil {
		return err
	}
	n.mu.Lock()
	n.gwv6 = net.IP{}
	n.mu.Unlock()
	return nil
}

// AddStaticRoute adds a static route to the sandbox.
func (n *Namespace) AddStaticRoute(r *types.StaticRoute) error {
	if err := n.programRoute(r.Destination, r.NextHop, true); err != nil {
		return err
	}
	n.mu.Lock()
	n.staticRoutes = append(n.staticRoutes, r)
	n.mu.Unlock()
	return nil
}

// RemoveStaticRoute removes a static route from the sandbox.
func (n *Namespace) RemoveStaticRoute(r *types.StaticRoute) error {
	if err := n.programRoute(r.Destination, r.NextHop, false); err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.staticRoutes = slices.DeleteFunc(n.staticRoutes, func(route *types.StaticRoute) bool {
		return route == r
	})
	return nil
}

// programRoute adds or removes a route to destination through nextHop. As the
// kernel does for a route without a device, it goes through the interface
// whose subnet has nextHop.
func (n *Namespace) programRoute(destination *net.IPNet, nextHop net.IP, add bool) error {
	if n.stack == nil {
		return n.notImplemented("changing the route to " + destination.String())
	}
	dstName, ok := n.interfaceReaching(nextHop)
	if !ok {
		return fmt.Errorf("network sandbox %s: no interface has a subnet with %s", n.path, nextHop)
	}
	if add {
		return n.stack.AddRoute(destination, nextHop, dstName)
	}
	return n.stack.RemoveRoute(destination, nextHop, dstName)
}

func (n *Namespace) interfaceReaching(ip net.IP) (string, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, i := range n.iFaces {
		for _, address := range append([]*net.IPNet{i.address, i.addressIPv6}, i.llAddrs...) {
			if address != nil && address.Contains(ip) {
				return i.dstName, true
			}
		}
	}
	return "", false
}

// SetDefaultRouteIPv4 sets up a connected route to 0.0.0.0 via the Interface
// with srcName.
func (n *Namespace) SetDefaultRouteIPv4(srcName string) error {
	if err := n.programDefaultRoute(srcName, defaultRouteIPv4, true); err != nil {
		return err
	}
	n.mu.Lock()
	n.defRoute4SrcName = srcName
	n.mu.Unlock()
	return nil
}

// SetDefaultRouteIPv6 sets up a connected route to [::] via the Interface
// with srcName.
func (n *Namespace) SetDefaultRouteIPv6(srcName string) error {
	if err := n.programDefaultRoute(srcName, defaultRouteIPv6, true); err != nil {
		return err
	}
	n.mu.Lock()
	n.defRoute6SrcName = srcName
	n.mu.Unlock()
	return nil
}

// UnsetDefaultRouteIPv4 unsets the previously set default IPv4 default route
// in the sandbox. It is a no-op if no gateway was set.
func (n *Namespace) UnsetDefaultRouteIPv4() error {
	n.mu.Lock()
	srcName := n.defRoute4SrcName
	n.mu.Unlock()
	if srcName == "" {
		return nil
	}
	if err := n.programDefaultRoute(srcName, defaultRouteIPv4, false); err != nil {
		return err
	}
	n.mu.Lock()
	n.defRoute4SrcName = ""
	n.mu.Unlock()
	return nil
}

// UnsetDefaultRouteIPv6 unsets the previously set default IPv6 default route
// in the sandbox. It is a no-op if no gateway was set.
func (n *Namespace) UnsetDefaultRouteIPv6() error {
	n.mu.Lock()
	srcName := n.defRoute6SrcName
	n.mu.Unlock()
	if srcName == "" {
		return nil
	}
	if err := n.programDefaultRoute(srcName, defaultRouteIPv6, false); err != nil {
		return err
	}
	n.mu.Lock()
	n.defRoute6SrcName = ""
	n.mu.Unlock()
	return nil
}

// programDefaultRoute adds or removes the default route without a next hop
// that the Interface with srcName has among its routes.
func (n *Namespace) programDefaultRoute(srcName string, destination *net.IPNet, add bool) error {
	if n.stack == nil {
		return n.notImplemented("changing the default route to " + srcName)
	}
	iface := n.InterfaceBySrcName(srcName)
	if iface == nil {
		if add {
			return fmt.Errorf("network sandbox %s: no interface %s", n.path, srcName)
		}
		// The interface has been removed along with its routes.
		return nil
	}
	if !slices.ContainsFunc(iface.routes, func(route *net.IPNet) bool { return route.String() == destination.String() }) {
		return fmt.Errorf("network sandbox %s: interface %s has no default route", n.path, srcName)
	}
	if add {
		return n.stack.AddRoute(destination, nil, iface.dstName)
	}
	return n.stack.RemoveRoute(destination, nil, iface.dstName)
}
