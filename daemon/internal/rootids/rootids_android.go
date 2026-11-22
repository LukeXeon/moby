// Package rootids returns the host identity that owns files of a container's
// root user.
package rootids

import (
	"os"

	"github.com/moby/sys/user"
)

// Pair returns the host UID and GID of the container's root user.
//
// An Android app cannot create user namespaces or chown files to other users.
// Containers run in gVisor, which presents files owned by the app's own UID
// and GID as owned by root, as rootless Docker's user namespace presents the
// invoking user. Without remapping, root is therefore the app's identity.
func Pair(m user.IdentityMapping) (uid, gid int) {
	if m.Empty() {
		return os.Getuid(), os.Getgid()
	}
	return m.RootPair()
}
