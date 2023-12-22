/*
*  Copyright (c) 2023 NetEase Inc.
*
*  Licensed under the Apache License, Version 2.0 (the "License");
*  you may not use this file except in compliance with the License.
*  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
*  Unless required by applicable law or agreed to in writing, software
*  distributed under the License is distributed on an "AS IS" BASIS,
*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*  See the License for the specific language governing permissions and
*  limitations under the License.
 */

/*
* Project: CurveAdm
* Created Date: 2023-12-13
* Author: liuminjian
 */

package module

import (
	"context"
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
	if cfg.Protocol == HTTP_PROTOCOL {
		client, err = NewHTTPClient(*cfg.GetHTTPConfig())
	} else {
		client, err = NewSSHClient(*cfg.GetSSHConfig())
	}
	return
}
