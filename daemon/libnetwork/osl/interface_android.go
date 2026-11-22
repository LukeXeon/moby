package osl

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/v2/daemon/libnetwork/types"
)

const (
	// AdvertiseAddrNMsgsMin defines the minimum number of ARP/NA messages sent when an
	// interface is configured.
	// Zero can be used to disable unsolicited ARP/NA.
	AdvertiseAddrNMsgsMin = 0
	// AdvertiseAddrNMsgsMax defines the maximum number of ARP/NA messages sent when an
	// interface is configured. It's three, to match RFC-5227 Section 1.1
	// ("PROBE_NUM=3") and RFC-4861 MAX_NEIGHBOR_ADVERTISEMENT.
	AdvertiseAddrNMsgsMax = 3

	// AdvertiseAddrIntervalMin defines the minimum interval between ARP/NA messages
	// sent when an interface is configured.
	AdvertiseAddrIntervalMin = 100 * time.Millisecond
	// AdvertiseAddrIntervalMax defines the maximum interval between ARP/NA messages
	// sent when an interface is configured. The max of 2s matches RFC-5227
	// PROBE_MAX.
	AdvertiseAddrIntervalMax = 2 * time.Second
)

// newInterface creates a new interface in the given namespace using the
// provided options.
func newInterface(ns *Namespace, srcName, dstPrefix, dstName string, options ...IfaceOption) (*Interface, error) {
	i := &Interface{
		srcName:   srcName,
		dstPrefix: dstPrefix,
		dstName:   dstName,
		ns:        ns,
	}
	for _, opt := range options {
		if opt != nil {
			if err := opt(i); err != nil {
				return nil, err
			}
		}
	}
	if ns.isDefault {
		i.dstName = i.srcName
	}
	return i, nil
}

// Interface represents the settings and identity of a network device.
type Interface struct {
	srcName               string
	dstPrefix             string
	dstName               string
	master                string
	mac                   net.HardwareAddr
	address               *net.IPNet
	addressIPv6           *net.IPNet
	llAddrs               []*net.IPNet
	routes                []*net.IPNet
	bridge                bool
	sysctls               []string
	advertiseAddrNMsgs    int
	advertiseAddrInterval time.Duration
	createdInContainer    bool
	ns                    *Namespace
}

// SrcName returns the name of the interface in the origin network namespace.
func (i *Interface) SrcName() string {
	return i.srcName
}

// DstName returns the final interface name in the target network namespace.
func (i *Interface) DstName() string {
	return i.dstName
}

// AddressIPv6 returns the IPv6 address for the interface.
func (i *Interface) AddressIPv6() *net.IPNet {
	return types.GetIPNetCopy(i.addressIPv6)
}

// Remove an interface from the sandbox.
func (i *Interface) Remove() error {
	return i.ns.RemoveInterface(i)
}

// Statistics returns the statistics of the interface.
func (i *Interface) Statistics() (*types.InterfaceStatistics, error) {
	return nil, i.ns.notImplemented("reading the statistics of " + i.dstName)
}

// AddInterface moves the link srcName into the network stack of the task,
// with the addresses and routes in options. Bridges and masters in the
// sandbox, which the overlay driver uses, links that a driver created in the
// sandbox, and interface sysctls are not supported.
func (n *Namespace) AddInterface(_ context.Context, srcName, dstPrefix, dstName string, options ...IfaceOption) error {
	if n.stack == nil {
		return n.notImplemented("adding interface " + srcName)
	}
	i, err := newInterface(n, srcName, dstPrefix, dstName, options...)
	if err != nil {
		return err
	}
	switch {
	case i.bridge || i.master != "":
		return n.notImplemented("adding a bridge or a bridge port " + srcName)
	case i.createdInContainer:
		return n.notImplemented("configuring interface " + srcName + " created in the sandbox")
	case len(i.sysctls) > 0:
		return n.notImplemented("setting " + strings.Join(i.sysctls, ", ") + " on " + srcName)
	}
	var addresses []*net.IPNet
	for _, address := range append([]*net.IPNet{i.address, i.addressIPv6}, i.llAddrs...) {
		if address != nil {
			addresses = append(addresses, address)
		}
	}

	// Names are chosen and recorded under the lock, as on Linux.
	n.mu.Lock()
	defer n.mu.Unlock()
	if i.dstName == "" {
		i.dstName = n.generateIfaceName(dstPrefix)
	}
	if err := n.stack.AddInterface(srcName, i.dstName, i.mac, addresses); err != nil {
		return fmt.Errorf("adding interface %s to %s as %s: %w", srcName, n.path, i.dstName, err)
	}
	for _, route := range i.routes {
		if route.IP.IsUnspecified() {
			// A default route is set if the interface becomes the gateway.
			continue
		}
		if err := n.stack.AddRoute(route, nil, i.dstName); err != nil {
			return errors.Join(fmt.Errorf("adding route %s to interface %s: %w", route, i.dstName, err),
				n.stack.RemoveInterface(i.dstName))
		}
	}
	n.iFaces = append(n.iFaces, i)
	return nil
}

// generateIfaceName returns prefix with the lowest number that no interface
// of the sandbox has, as on Linux.
func (n *Namespace) generateIfaceName(prefix string) string {
	used := make(map[int]bool)
	for _, i := range n.iFaces {
		if s, ok := strings.CutPrefix(i.dstName, prefix); ok {
			if v, err := strconv.Atoi(s); err == nil && v >= 0 && s != "-0" {
				used[v] = true
			}
		}
	}
	suffix := 0
	for used[suffix] {
		suffix++
	}
	return prefix + strconv.Itoa(suffix)
}

// RemoveInterface removes the interface from the sandbox. It may be called
// more than once for the same Interface.
func (n *Namespace) RemoveInterface(i *Interface) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if !slices.Contains(n.iFaces, i) {
		return nil
	}
	if n.stack != nil {
		if err := n.stack.RemoveInterface(i.dstName); err != nil {
			return fmt.Errorf("removing interface %s from %s: %w", i.dstName, n.path, err)
		}
	}
	n.iFaces = slices.DeleteFunc(n.iFaces, func(iface *Interface) bool {
		return iface == i
	})
	return nil
}
