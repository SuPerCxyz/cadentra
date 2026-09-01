package agent

import "syscall"

type statfsT = syscall.Statfs_t

func statfs(path string, buf *syscall.Statfs_t) error {
	return syscall.Statfs(path, buf)
}
