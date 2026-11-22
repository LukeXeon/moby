package local

import "os"

var removeAllFn = os.RemoveAll

// SetRemoveAll sets the function that removes volume directories. It must be
// called before the daemon starts.
//
// On Android, containers write volumes through the embedder's filesystem layer,
// which may leave read-only directories that an app without CAP_DAC_OVERRIDE
// cannot empty with os.RemoveAll, as well as emulated hardlink objects that
// os.RemoveAll would leak.
func SetRemoveAll(fn func(string) error) {
	removeAllFn = fn
}

func removeAll(path string) error {
	return removeAllFn(path)
}
