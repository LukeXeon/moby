package osl

import (
	"context"
	"net"
	"slices"
	"sync"

	"github.com/moby/moby/v2/daemon/libnetwork/internal/nftables"
	"github.com/moby/moby/v2/daemon/libnetwork/types"
)

// On android, the network stack of a container is the netstack of the Sentry
// that runs it. An app can neither create nor enter a kernel network
// namespace, and seccomp kills it on mount and umount2, so a Namespace holds
// no kernel object: it records the configuration of the sandbox, which starts
// with only the loopback interface of the Sentry. The default Namespace stands
// for the network of the daemon's own process. A change that has to be applied
// to the network stack of a Sentry fails with a NotImplemented error.

// NewSandbox returns the Namespace with the given key. osCreate is false for
// the default Namespace.
func NewSandbox(key string, osCreate, isRestore bool) (*Namespace, error) {
	return &Namespace{path: key, isDefault: !osCreate}, nil
}

// GetSandboxForExternalKey returns the Namespace of a container whose runtime
// created its network stack. On Linux basePath is the network namespace of the
// runtime's process, which a Sentry does not have, so it is not used.
func GetSandboxForExternalKey(basePath string, key string) (*Namespace, error) {
	return &Namespace{path: key}, nil
}

// Namespace represents the network stack of a sandbox.
type Namespace struct {
	path             string // path is the key of the sandbox. It is safe to access it concurrently.
	isDefault        bool   // isDefault is true when Namespace represents the daemon's network. It is safe to access it concurrently.
	iFaces           []*Interface
	gw               net.IP
	gwv6             net.IP
	defRoute4SrcName string
	defRoute6SrcName string
	staticRoutes     []*types.StaticRoute
	mu               sync.Mutex
}

// notImplemented returns the error of a change that the network stack of the
// sandbox does not support.
func (n *Namespace) notImplemented(change string) error {
	return types.NotImplementedErrorf("network sandbox %s: %s is not supported on android", n.path, change)
}

// Interfaces returns the collection of Interface previously added with the
// AddInterface method or restored.
func (n *Namespace) Interfaces() []*Interface {
	n.mu.Lock()
	defer n.mu.Unlock()
	return slices.Clone(n.iFaces)
}

// InterfaceBySrcName returns a pointer to the Interface with a matching srcName, else nil.
func (n *Namespace) InterfaceBySrcName(srcName string) *Interface {
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, iface := range n.iFaces {
		if iface.srcName == srcName {
			return iface
		}
	}
	return nil
}

// GetLoopbackIfaceName returns the name of the loopback interface
func (n *Namespace) GetLoopbackIfaceName() string {
	return "lo"
}

// AddAliasIP adds the passed IP address to the named interface
func (n *Namespace) AddAliasIP(ifName string, ip *net.IPNet) error {
	return n.notImplemented("adding alias address " + ip.String() + " to " + ifName)
}

// RemoveAliasIP removes the passed IP address from the named interface
func (n *Namespace) RemoveAliasIP(ifName string, ip *net.IPNet) error {
	return n.notImplemented("removing alias address " + ip.String() + " from " + ifName)
}

// DisableARPForVIP disables ARP replies and requests for VIP addresses
// on a particular interface.
func (n *Namespace) DisableARPForVIP(srcName string) error {
	return n.notImplemented("disabling ARP on " + srcName)
}

// InvokeFunc runs f in the network of the sandbox. Only the default Namespace,
// the daemon's own network, can run code.
func (n *Namespace) InvokeFunc(f func()) error {
	if !n.isDefault {
		return n.notImplemented("running a function in the network namespace")
	}
	f()
	return nil
}

// Key returns the key of the sandbox.
func (n *Namespace) Key() string {
	return n.path
}

// Destroy destroys the sandbox. A Namespace holds no resources.
func (n *Namespace) Destroy() error {
	return nil
}

// RestoreInterfaces restores the interfaces of the sandbox from the
// configuration in the store.
func (n *Namespace) RestoreInterfaces(_ context.Context, interfaces map[Iface][]IfaceOption) error {
	for iface, opts := range interfaces {
		i, err := newInterface(n, iface.SrcName, iface.DstPrefix, iface.DstName, opts...)
		if err != nil {
			return err
		}
		if n.isDefault {
			continue
		}
		n.mu.Lock()
		n.iFaces = append(n.iFaces, i)
		n.mu.Unlock()
	}
	return nil
}

func (n *Namespace) RestoreRoutes(routes []*types.StaticRoute) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.staticRoutes = append(n.staticRoutes, routes...)
}

func (n *Namespace) RestoreGateway(ipv4 bool, gw net.IP, srcName string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if gw == nil {
		// There's no gateway address, so the default route is bound to the interface.
		if ipv4 {
			n.defRoute4SrcName = srcName
		} else {
			n.defRoute6SrcName = srcName
		}
		return
	}

	if ipv4 {
		n.gw = gw
	} else {
		n.gwv6 = gw
	}
}

// IPv6LoEnabled returns true: the loopback interface of a Sentry always has
// the address '::1'.
func (n *Namespace) IPv6LoEnabled() bool {
	return true
}

// RefreshIPv6LoEnabled does nothing, see [Namespace.IPv6LoEnabled].
func (n *Namespace) RefreshIPv6LoEnabled() {}

// ApplyOSTweaks does nothing. The kernel knobs are for the sandboxes of swarm
// load balancers and ingress, which need the overlay driver.
func (n *Namespace) ApplyOSTweaks([]SandboxType) {}

// ApplyNFTable applies changes to the specified nftables table in the namespace.
func (n *Namespace) ApplyNFTable(_ context.Context, family nftables.Family, name string, _ func(*nftables.Modifier) error, _ ...nftables.Changes) error {
	return n.notImplemented("applying nftables table " + string(family) + " " + name)
}
