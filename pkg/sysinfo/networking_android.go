package sysinfo

// applyNetworkingInfo adds networking information to the info. On android the
// traffic of containers does not pass through the kernel's IP forwarding: the
// embedder of the daemon routes it in user space, so the host's sysctl does not
// apply.
func applyNetworkingInfo(info *SysInfo) {
	info.IPv4ForwardingDisabled = false
}
