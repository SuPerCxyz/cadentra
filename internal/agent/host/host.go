package host

import (
	"context"
	"io"
	"os"
)

// HostAdapter 宿主机抽象接口
// NativeHostAdapter 使用 HOST_ROOT=/
// ContainerHostAdapter 使用 HOST_ROOT=/host
type HostAdapter interface {
	// HostRoot 返回映射根
	HostRoot() string
	// MapPath 将逻辑路径映射到实际操作路径
	MapPath(logicalPath string) string
	ValidatePath(logicalPath string) error

	// 文件操作
	WriteFile(ctx context.Context, logicalPath string, data []byte, mode os.FileMode) error
	AtomicReplace(ctx context.Context, logicalPath string, data []byte, mode os.FileMode) error
	Chmod(ctx context.Context, logicalPath string, mode os.FileMode) error
	Chown(ctx context.Context, logicalPath, user, group string) error
	MkdirAll(ctx context.Context, logicalPath string, mode os.FileMode) error
	Remove(ctx context.Context, logicalPath string) error
	ReadFile(ctx context.Context, logicalPath string) ([]byte, error)
	OpenRead(ctx context.Context, logicalPath string) (io.ReadCloser, os.FileInfo, error)
	AtomicReplaceReader(ctx context.Context, logicalPath string, r io.Reader, mode os.FileMode) error
	Stat(ctx context.Context, logicalPath string) (os.FileInfo, error)

	// systemd 操作
	InstallUnit(ctx context.Context, unitName, content string) error
	UpdateUnit(ctx context.Context, unitName, content string) error
	DaemonReload(ctx context.Context) error
	EnableService(ctx context.Context, unitName string) error
	DisableService(ctx context.Context, unitName string) error
	StartService(ctx context.Context, unitName string) error
	StopService(ctx context.Context, unitName string) error
	RestartService(ctx context.Context, unitName string) error
	ServiceStatus(ctx context.Context, unitName string) (string, error)

	// 执行
	RunCommand(ctx context.Context, name string, args ...string) (string, error)
	RunCommandWithEnv(ctx context.Context, name string, env []string, args ...string) (string, error)
}
