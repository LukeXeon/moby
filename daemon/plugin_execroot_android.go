package daemon

import (
	"path/filepath"

	"github.com/moby/moby/v2/daemon/config"
)

// An Android app can only write its own data directories, so plugin execution
// state lives under the daemon's exec-root. Plugin socket paths must still fit
// in 108 bytes.
func getPluginExecRoot(cfg *config.Config) string {
	return filepath.Join(cfg.ExecRoot, "plugins")
}
