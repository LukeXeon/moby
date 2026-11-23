package resolvconf

// An Android system has no /etc/resolv.conf. The network that the daemon sees
// belongs to its embedder, which writes the resolver configuration of that
// network to a file and names it with SetPath before the daemon starts, as
// RootlessKit gives rootless Docker the resolv.conf of its network namespace.
var path string

// SetPath sets the path that Path returns.
func SetPath(p string) {
	path = p
}

// Path returns the path to the resolv.conf of the daemon's network.
func Path() string {
	return path
}
