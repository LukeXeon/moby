//go:build !android

package snapshotter

import (
	"github.com/moby/locker"
	"github.com/moby/moby/v2/daemon/internal/mountref"
	"github.com/moby/sys/user"
)

func newMounter(home string, snapshotter string, idMap user.IdentityMapping) Mounter {
	return &refCountMounter{
		base: mounter{
			home:        home,
			snapshotter: snapshotter,
			idMap:       idMap,
		},
		rc:     mountref.NewCounter(isMounted),
		locker: locker.New(),
	}
}
