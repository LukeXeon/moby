//go:build (linux && !android) || freebsd

package daemon

// hostResolvConf returns the host's resolv.conf, which containers in the
// host's network and the embedded resolver use as it is.
func hostResolvConf() string {
	return "/etc/resolv.conf"
}
