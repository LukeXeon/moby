//go:build !android

package local

import "os"

func removeAll(path string) error {
	return os.RemoveAll(path)
}
