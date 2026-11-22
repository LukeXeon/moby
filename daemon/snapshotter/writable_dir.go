package snapshotter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/containerd/containerd/v2/core/mount"
	cerrdefs "github.com/containerd/errdefs"
)

// WritableDir returns the directory that holds the writable layer of an
// overlayfs snapshot, given the snapshot's mounts: the source of a bind mount
// for a snapshot without parents, or the upper directory of an overlay mount.
// It is how the daemon writes to a snapshot where it cannot mount one.
func WritableDir(mounts []mount.Mount) (string, error) {
	if len(mounts) != 1 {
		return "", fmt.Errorf("writable directory of %d mounts: %w", len(mounts), cerrdefs.ErrNotImplemented)
	}
	m := mounts[0]
	switch m.Type {
	case "bind":
		if slices.Contains(m.Options, "ro") {
			return "", fmt.Errorf("writable directory of a read-only bind mount of %s: %w", m.Source, cerrdefs.ErrInvalidArgument)
		}
		return m.Source, nil
	case "overlay":
		for _, option := range m.Options {
			if dir, ok := strings.CutPrefix(option, "upperdir="); ok {
				return dir, nil
			}
		}
		return "", fmt.Errorf("writable directory of a read-only overlay mount: %w", cerrdefs.ErrInvalidArgument)
	default:
		return "", fmt.Errorf("writable directory of a %q mount: %w", m.Type, cerrdefs.ErrNotImplemented)
	}
}
