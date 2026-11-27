package daemon

import (
	"context"

	"github.com/moby/moby/api/types/system"
)

// On Android, containers run in the embedder's sandbox rather than under a
// host OCI runtime binary, and docker-init is a guest executable from the
// app's runtime assets. Executing host processes from Go also invokes
// pidfd_open(2), which Android 10's seccomp filter terminates with SIGSYS.

func populateRuncCommit(*system.Commit, *configStore) error {
	return nil
}

func (*Daemon) populateInitCommit(_ context.Context, v *system.Info, cfg *configStore) error {
	v.InitBinary = cfg.GetInitPath()
	return nil
}

func populateRuncVersion(*configStore, *system.VersionResponse) error {
	return nil
}

func populateInitVersion(context.Context, *configStore, *system.VersionResponse) error {
	return nil
}
