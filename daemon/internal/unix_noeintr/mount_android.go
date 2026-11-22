package unix_noeintr

import "golang.org/x/sys/unix"

// The seccomp filter of Android app processes kills the process on mount(2)
// and umount2(2) instead of failing them. An app process therefore has no
// mount of its own: Mount fails with EPERM, as it does for an unprivileged
// process, and Unmount fails with EINVAL, as for a target that is not a mount
// point.

func Mount(source string, target string, fstype string, flags uintptr, data string) error {
	return unix.EPERM
}

func Unmount(target string, flags int) error {
	return unix.EINVAL
}
