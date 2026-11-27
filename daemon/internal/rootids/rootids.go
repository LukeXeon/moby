//go:build !android

// Package rootids returns the host identity that owns files of a container's
// root user.
package rootids

import "github.com/moby/sys/user"

// Pair returns the host UID and GID of the container's root user.
func Pair(m user.IdentityMapping) (uid, gid int) {
	return m.RootPair()
}
