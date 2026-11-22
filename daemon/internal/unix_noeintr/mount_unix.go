//go:build !darwin && !windows && !android

package unix_noeintr

import "golang.org/x/sys/unix"

func Mount(source string, target string, fstype string, flags uintptr, data string) (err error) {
	Retry(func() error {
		err = unix.Mount(source, target, fstype, flags, data)
		return err
	})
	return err
}

func Unmount(target string, flags int) (err error) {
	Retry(func() error {
		err = unix.Unmount(target, flags)
		return err
	})
	return err
}
