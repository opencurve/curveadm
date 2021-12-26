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

package module

import (
	"errors"
	"fmt"
	"net"

	"github.com/melbahja/goph"
	"github.com/opencurve/curveadm/pkg/log"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	User           string
	Host           string
	Port           uint
	PrivateKeyPath string
	Timeout        int
}

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

func NewSshClient(sshConfig SSHConfig) (*goph.Client, error) {
	user := sshConfig.User
	host := sshConfig.Host
	port := sshConfig.Port
	privateKeyPath := sshConfig.PrivateKeyPath

	auth, err := goph.Key(privateKeyPath, "")
	if err != nil {
		log.Error("SSHAuth",
			log.Field("PrivateKeyPath", privateKeyPath),
			log.Field("error", err))
		return nil, err
	}

	client, err := goph.NewConn(&goph.Config{
		User:     user,
		Addr:     host,
		Port:     port,
		Auth:     auth,
		Callback: VerifyHost,
	})

	log.SwitchLevel(err)("SSHConnect",
		log.Field("user", user),
		log.Field("addr", fmt.Sprintf("%s:%d", host, port)),
		log.Field("PrivateKeyPath", privateKeyPath),
		log.Field("error", err))

	return client, err
}
