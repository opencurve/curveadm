package module

import (
	"context"
	"github.com/opencurve/curveadm/internal/errno"
)

type ConnectConfig struct {
	User              string
	Host              string
	SSHPort           uint
	HTTPPort          uint
	ForwardAgent      bool // ForwardAgent > PrivateKeyPath > Password
	BecomeMethod      string
	BecomeFlags       string
	BecomeUser        string
	PrivateKeyPath    string
	ConnectRetries    int
	ConnectTimeoutSec int
	Protocol          string
}

func (c *ConnectConfig) GetSSHConfig() *SSHConfig {
	return &SSHConfig{
		User:              c.User,
		Host:              c.Host,
		Port:              c.SSHPort,
		ForwardAgent:      c.ForwardAgent,
		BecomeMethod:      c.BecomeMethod,
		BecomeFlags:       c.BecomeFlags,
		BecomeUser:        c.BecomeUser,
		PrivateKeyPath:    c.PrivateKeyPath,
		ConnectRetries:    c.ConnectRetries,
		ConnectTimeoutSec: c.ConnectTimeoutSec,
	}
}

func (c *ConnectConfig) GetHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Host: c.Host,
		Port: c.HTTPPort,
	}
}

type RemoteClient interface {
	Protocol() string
	WrapperCommand(command string, execInLocal bool) (wrapperCmd string)
	RunCommand(ctx context.Context, command string) (out []byte, err error)
	RemoteAddr() (addr string)
	Upload(localPath string, remotePath string) (err error)
	Download(remotePath string, localPath string) (err error)
	Close()
}

func NewRemoteClient(cfg *ConnectConfig) (client RemoteClient, err error) {
	if cfg == nil {
		return
	}

	if cfg.Protocol == SSH_PROTOCOL {
		client, err = NewSSHClient(*cfg.GetSSHConfig())
		if err != nil {
			return nil, errno.ERR_SSH_CONNECT_FAILED.E(err)
		}
		return client, nil
	} else if cfg.Protocol == HTTP_PROTOCOL {
		client, err = NewHTTPClient(*cfg.GetHTTPConfig())
		if err != nil {
			return nil, errno.ERR_HTTP_CONNECT_FAILED.E(err)
		}
		return
	}
	return
}
