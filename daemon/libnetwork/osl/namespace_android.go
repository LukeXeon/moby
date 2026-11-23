package osl

import (
	"context"
	"net"
	"net/netip"
	"slices"
	"sync"

	"github.com/moby/moby/v2/daemon/libnetwork/internal/nftables"
	"github.com/moby/moby/v2/daemon/libnetwork/types"
)

// On android, the network stack of a container is the netstack of the Sentry
// that runs it. An app can neither create nor enter a kernel network
// namespace, and seccomp kills it on mount and umount2, so a Namespace holds
// no kernel object. The Namespace of a task changes the network stack of the
// task through the Network that the embedder of the daemon provides; other
// Namespaces only record their configuration. The default Namespace stands for
// the network of the daemon's own process.

// Network changes the network stacks of tasks for the daemon on android, as
// netlink does on Linux. The embedder sets it with SetNetwork before the
// daemon starts.
type Network interface {
	// Namespace returns the network namespace that path names, the
	// "/proc/<pid>/ns/net" of the process that a runtime created for a task.
	Namespace(path string) (NetworkNamespace, error)
}

// NetworkNamespace is the network stack of one task. Interface names are the
// names in the namespace.
type NetworkNamespace interface {
	// AddInterface moves the link srcName, which a network driver created in
	// the embedder's network, into the namespace as dstName.
	AddInterface(srcName, dstName string, mac net.HardwareAddr, addresses []*net.IPNet) error
	// RemoveInterface returns the interface dstName to the embedder's network
	// under its original name.
	RemoveInterface(dstName string) error
	// AddRoute adds a route through dstName. gateway is nil for a route
	// without a next hop.
	AddRoute(destination *net.IPNet, gateway net.IP, dstName string) error
	RemoveRoute(destination *net.IPNet, gateway net.IP, dstName string) error
	// Redirect makes what the namespace sends to from over network, "tcp" or
	// "udp", reach to in the daemon's network, as a DNAT rule in the
	// namespace to a service of the host does. It replaces the redirect of
	// the same network and address.
	Redirect(network string, from, to netip.AddrPort) error
	// Dial connects to address from the namespace, as a process in it does.
	Dial(network, address string) (net.Conn, error)
}

var network Network

// SetNetwork sets the Network that configures the network namespaces of tasks.
func SetNetwork(n Network) {
	network = n
}

// NewSandbox returns the Namespace with the given key. osCreate is false for
// the default Namespace.
func NewSandbox(key string, osCreate, isRestore bool) (*Namespace, error) {
	return &Namespace{path: key, isDefault: !osCreate}, nil
}

// GetSandboxForExternalKey returns the Namespace of a task whose runtime
// created its network stack. basePath is the network namespace of the task's
// process.
func GetSandboxForExternalKey(basePath string, key string) (*Namespace, error) {
	if network == nil {
		return nil, types.NotImplementedErrorf("network sandbox %s: the daemon has no network for tasks", key)
	}
	stack, err := network.Namespace(basePath)
	if err != nil {
		return nil, err
	}
	return &Namespace{path: key, stack: stack}, nil
}

// Namespace represents the network stack of a sandbox.
type Namespace struct {
	path             string           // path is the key of the sandbox. It is safe to access it concurrently.
	isDefault        bool             // isDefault is true when Namespace represents the daemon's network. It is safe to access it concurrently.
	stack            NetworkNamespace // stack is the network stack of a task, nil for other sandboxes. It is safe to access it concurrently.
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

// InvokeFunc runs f in the daemon's own network: the daemon cannot enter the
// network stack of a task, so the sockets that f opens are the daemon's. The
// embedded resolver, which opens its listeners this way, reaches the task
// through Redirect and Dial.
func (n *Namespace) InvokeFunc(f func()) error {
	f()
	return nil
}

// Redirect makes what the sandbox sends to from over network reach to in the
// daemon's network.
func (n *Namespace) Redirect(network string, from, to netip.AddrPort) error {
	if n.stack == nil {
		return n.notImplemented("redirecting " + network + " " + from.String())
	}
	return n.stack.Redirect(network, from, to)
}

// Dial connects to address from the sandbox.
func (n *Namespace) Dial(network, address string) (net.Conn, error) {
	if n.stack == nil {
		return nil, n.notImplemented("connecting to " + network + " " + address)
	}
	return n.stack.Dial(network, address)
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
