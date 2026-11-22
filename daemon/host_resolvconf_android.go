package daemon

import "github.com/moby/moby/v2/daemon/libnetwork/resolvconf"

// hostResolvConf returns the resolv.conf that the embedder names for the
// daemon's network: an Android system has no /etc/resolv.conf.
func hostResolvConf() string {
	return resolvconf.Path()
}
