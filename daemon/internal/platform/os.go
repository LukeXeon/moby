//go:build !android

package platform

import "runtime"

// OS is the operating system of the containers that the daemon runs, and so
// of the images it uses.
const OS = runtime.GOOS
