/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-08-01
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package hosts

import (
	comm "github.com/opencurve/curveadm/internal/configure/common"
	"github.com/opencurve/curveadm/internal/configure/curveadm"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
)

func (hc *HostConfig) get(i *comm.Item) interface{} {
	if v, ok := hc.config[i.Key()]; ok {
		return v
	}

	defaultValue := i.DefaultValue()
	if defaultValue != nil && utils.IsFunc(defaultValue) {
		return defaultValue.(func(*HostConfig) interface{})(hc)
	}
	return defaultValue
}

func (hc *HostConfig) getString(i *comm.Item) string {
	v := hc.get(i)
	if v == nil {
		return ""
	}
	return v.(string)
}

func (hc *HostConfig) getInt(i *comm.Item) int {
	v := hc.get(i)
	if v == nil {
		return 0
	}
	return v.(int)
}

func (hc *HostConfig) getBool(i *comm.Item) bool {
	v := hc.get(i)
	if v == nil {
		return false
	}
	return v.(bool)
}

func (hc *HostConfig) GetHost() string           { return hc.getString(CONFIG_HOST) }
func (hc *HostConfig) GetHostname() string       { return hc.getString(CONFIG_HOSTNAME) }
func (hc *HostConfig) GetSSHHostname() string    { return hc.getString(CONFIG_SSH_HOSTNAME) }
func (hc *HostConfig) GetSSHPort() int           { return hc.getInt(CONFIG_SSH_PORT) }
func (hc *HostConfig) GetPrivateKeyFile() string { return hc.getString(CONFIG_PRIVATE_CONFIG_FILE) }
func (hc *HostConfig) GetForwardAgent() bool     { return hc.getBool(CONFIG_FORWARD_AGENT) }
func (hc *HostConfig) GetBecomeUser() string     { return hc.getString(CONFIG_BECOME_USER) }
func (hc *HostConfig) GetEnvs() []string         { return hc.envs }

func (hc *HostConfig) GetLabels() []string {
	if len(hc.labels) == 0 {
		return []string{hc.GetHost()}
	}
	return hc.labels
}

func (hc *HostConfig) GetUser() string {
	user := hc.getString(CONFIG_USER)
	if user == "${user}" {
		return utils.GetCurrentUser()
	}
	return user
}

func (hc *HostConfig) GetSSHConfig() *module.SSHConfig {
	hostname := hc.GetSSHHostname()
	if len(hostname) == 0 {
		hostname = hc.GetHostname()
	}
	return &module.SSHConfig{
		User:              hc.GetUser(),
		Host:              hostname,
		Port:              (uint)(hc.GetSSHPort()),
		PrivateKeyPath:    hc.GetPrivateKeyFile(),
		ForwardAgent:      hc.GetForwardAgent(),
		BecomeMethod:      "sudo",
		BecomeFlags:       "-iu",
		BecomeUser:        hc.GetBecomeUser(),
		ConnectTimeoutSec: curveadm.GlobalCurveAdmConfig.GetSSHTimeout(),
		ConnectRetries:    curveadm.GlobalCurveAdmConfig.GetSSHRetries(),
	}
}
