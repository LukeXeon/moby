package buildkit

import (
	"errors"

	"github.com/opencontainers/runtime-spec/specs-go"
)

const networkName = "bridge"

// Build steps do not run on Android yet: newExecutor returns the stub
// executor, so no network namespace of a build step is set up.
func (iface *lnInterface) Set(*specs.Spec) error {
	return errors.New("buildkit network namespaces are not implemented on android")
}
