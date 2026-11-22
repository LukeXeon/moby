//go:build !android

package sysinfo

// applyNetworkingInfo adds networking information to the info.
func applyNetworkingInfo(info *SysInfo) {
	info.IPv4ForwardingDisabled = !readProcBool("/proc/sys/net/ipv4/ip_forward")
}
