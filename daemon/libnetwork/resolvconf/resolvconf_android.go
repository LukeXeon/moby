package resolvconf

import "github.com/moby/moby/v2/daemon/libnetwork/internal/resolvconf"

// SetPath is an alias for [resolvconf.SetPath], which is internal to
// libnetwork, for the embedder of the daemon.
func SetPath(path string) {
	resolvconf.SetPath(path)
}
