package osl

import (
	"context"
	"net"
	"slices"
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

// AddInterface adds an existing network interface to the sandbox.
func (n *Namespace) AddInterface(_ context.Context, srcName, _, _ string, _ ...IfaceOption) error {
	return n.notImplemented("adding interface " + srcName)
}

// RemoveInterface removes the interface from the sandbox. It may be called
// more than once for the same Interface.
func (n *Namespace) RemoveInterface(i *Interface) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.iFaces = slices.DeleteFunc(n.iFaces, func(iface *Interface) bool {
		return iface == i
	})
	return nil
}
