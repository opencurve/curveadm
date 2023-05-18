/*
 *  Copyright (c) 2021 NetEase Inc.
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
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package module

import (
	"errors"
	"github.com/opencurve/curveadm/pkg/log/zaplog"
	"go.uber.org/zap"
	"net"
	"time"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

type (
	SSHConfig struct {
		User              string
		Host              string
		Port              uint
		ForwardAgent      bool // ForwardAgent > PrivateKeyPath > Password
		BecomeMethod      string
		BecomeFlags       string
		BecomeUser        string
		PrivateKeyPath    string
		ConnectRetries    int
		ConnectTimeoutSec int
	}

	SSHClient struct {
		client *goph.Client
		config SSHConfig
	}
)

func askIsHostTrusted(host string, key ssh.PublicKey) bool {
	//	format := "Unknown Host: %s \\nFingerprint: %s \\nWould you likt to add it?[y/N]: "
	//	prompt := fmt.Sprintf(format, host, ssh.FingerprintSHA256(key))
	//	return tui.ConfirmYes(prompt)
	return true
}

func VerifyHost(host string, remote net.Addr, key ssh.PublicKey) error {
	hostFound, err := goph.CheckKnownHost(host, remote, key, "")

	/*
	 * Host in known hosts but key mismatch!
	 * Maybe because of MAN IN THE MIDDLE ATTACK!
	 */
	if hostFound && err != nil {
		return err
	} else if hostFound && err == nil { // handshake because public key already exists.
		return nil
	} else if askIsHostTrusted(host, key) == false { // Ask user to check if he trust the host public key.
		// Make sure to return error on non trusted keys.
		return errors.New("you typed no, aborted!")
	}

	// Add the new host to known hosts file.
	return goph.AddKnownHost(host, remote, key, "")
}

func (client *SSHClient) Client() *goph.Client {
	return client.client
}

func (client *SSHClient) Config() SSHConfig {
	return client.config
}

func NewSSHClient(config SSHConfig) (*SSHClient, error) {
	user := config.User
	host := config.Host
	port := config.Port
	forwardAgent := config.ForwardAgent
	privateKeyPath := config.PrivateKeyPath
	connTimeoutSec := config.ConnectTimeoutSec
	maxRetries := config.ConnectRetries

	var auth goph.Auth
	var err error
	if forwardAgent {
		auth, err = goph.UseAgent()
	} else {
		auth, err = goph.Key(privateKeyPath, "")
	}

	if err != nil {
		zaplog.Error("Create SSH auth",
			zap.String("user", user),
			zap.String("host", host),
			zap.Uint("port", port),
			zap.Bool("forwardAgent", forwardAgent),
			zap.String("privateKeyPath", privateKeyPath),
			zap.Any("error", err))
		return nil, err
	}

	tries := 0
connect:
	tries++
	client, err := goph.NewConn(&goph.Config{
		User:     user,
		Addr:     host,
		Port:     port,
		Auth:     auth,
		Timeout:  time.Duration(connTimeoutSec) * time.Second,
		Callback: VerifyHost,
	})

	zaplog.Error("Connect remote SSH",
		zap.String("user", user),
		zap.String("host", host),
		zap.Uint("port", port),
		zap.Bool("forwardAgent", forwardAgent),
		zap.String("privateKeyPath", privateKeyPath),
		zap.Int("timeoutSec", connTimeoutSec),
		zap.Int("maxRetries", maxRetries),
		zap.Int("tries", tries),
		zap.Any("error", err))

	if err != nil {
		if tries < maxRetries {
			goto connect
		}
	}

	return &SSHClient{
		client: client,
		config: config,
	}, err
}
