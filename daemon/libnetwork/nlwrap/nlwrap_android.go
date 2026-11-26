// Package nlwrap reports netlink as unavailable on android without calling into
// vishvananda/netlink. An Android app's SELinux policy denies bind(2) on
// AF_NETLINK sockets; vishvananda/netlink's getNetlinkSocket wraps the socket
// in an *os.File before bind and then calls unix.Close(fd) on bind failure,
// leaving an unclosed *os.File whose GC finalizer later closes whatever live
// file descriptor reuses that number.
package nlwrap

import (
	"syscall"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type Handle struct {
	*netlink.Handle
}

func NewHandle(...int) (Handle, error) {
	return Handle{}, syscall.EACCES
}

func NewHandleAt(netns.NsHandle, ...int) (Handle, error) {
	return Handle{}, syscall.EACCES
}

func (nlh Handle) Close() {
	if nlh.Handle != nil {
		nlh.Handle.Close()
	}
}

func (Handle) AddrList(netlink.Link, int) ([]netlink.Addr, error) {
	return nil, syscall.EACCES
}

func AddrList(netlink.Link, int) ([]netlink.Addr, error) {
	return nil, syscall.EACCES
}

func (Handle) ConntrackDeleteFilters(netlink.ConntrackTableType, netlink.InetFamily, ...netlink.CustomConntrackFilter) (uint, error) {
	return 0, syscall.EACCES
}

func ConntrackTableList(netlink.ConntrackTableType, netlink.InetFamily) ([]*netlink.ConntrackFlow, error) {
	return nil, syscall.EACCES
}

func (Handle) LinkByName(string) (netlink.Link, error) {
	return nil, syscall.EACCES
}

func LinkByName(string) (netlink.Link, error) {
	return nil, syscall.EACCES
}

func (Handle) LinkList() ([]netlink.Link, error) {
	return nil, syscall.EACCES
}

func LinkList() ([]netlink.Link, error) {
	return nil, syscall.EACCES
}

func LinkSubscribeWithOptions(chan<- netlink.LinkUpdate, <-chan struct{}, netlink.LinkSubscribeOptions) error {
	return syscall.EACCES
}

func (Handle) RouteList(netlink.Link, int) ([]netlink.Route, error) {
	return nil, syscall.EACCES
}

func (Handle) XfrmPolicyList(int) ([]netlink.XfrmPolicy, error) {
	return nil, syscall.EACCES
}

func (Handle) XfrmStateList(int) ([]netlink.XfrmState, error) {
	return nil, syscall.EACCES
}
