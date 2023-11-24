package module

import (
	"context"
)

type RemoteClient interface {
	Protocol() string
	WrapperCommand(command string, execInLocal bool) (wrapperCmd string)
	RunCommand(ctx context.Context, command string) (out []byte, err error)
	RemoteAddr() (addr string)
	Upload(localPath string, remotePath string) (err error)
	Download(remotePath string, localPath string) (err error)
	Close()
}
