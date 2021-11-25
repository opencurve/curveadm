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

package configure

import (
	"fmt"
	"strconv"

	"github.com/opencurve/curveadm/internal/log"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	ROLE_ETCD       string = "etcd"
	ROLE_MDS        string = "mds"
	ROLE_METASERVER string = "metaserver"

	KEY_USER               string = "user"
	KEY_SSH_PORT           string = "ssh_port"
	KEY_PRIVATE_KEY_FILE   string = "private_key_file"
	KEY_CONTAINER_IMAGE    string = "container_image"
	KEY_LOG_DIR            string = "log_dir"
	KEY_DATA_DIR           string = "data_dir"
	KEY_LISTEN_IP          string = "listen.ip"
	KEY_LISTEN_PORT        string = "listen.port"
	KEY_LISTEN_CLIENT_PORT string = "listen.client_port" // etcd
	KEY_LISTEN_DUMMY_PORT  string = "listen.dummy_port"  // mds
	KEY_LISTEN_EXTERNAL_IP string = "listen.external_ip" // metaserver
	KEY_VARIABLE           string = "variable"

	DEFAULT_CURVEFS_DIR             = "/usr/local/curvefs"
	DEFAULT_SSH_PORT                = 22
	DEFAULT_ETCD_LISTEN_PEER_PORT   = 2380
	DEFAULT_ETCD_LISTEN_CLIENT_PORT = 2379
	DEFAULT_MDS_LISTEN_PORT         = 6700
	DEFAULT_MDS_LISTEN_DUMMY_PORT   = 7700
	DEFAULT_METASERVER_LISTN_PORT   = 6701
)

var (
	excludeServiceConfig = map[string]bool{
		KEY_USER:               true,
		KEY_SSH_PORT:           true,
		KEY_PRIVATE_KEY_FILE:   true,
		KEY_CONTAINER_IMAGE:    true,
		KEY_DATA_DIR:           true,
		KEY_LOG_DIR:            true,
		KEY_LISTEN_IP:          true,
		KEY_LISTEN_PORT:        true,
		KEY_LISTEN_CLIENT_PORT: true,
		KEY_LISTEN_DUMMY_PORT:  true,
		KEY_LISTEN_EXTERNAL_IP: true,
		KEY_VARIABLE:           true,
	}
)

type SshConfig struct {
	User           string
	Host           string
	Port           uint
	PrivateKeyPath string
	Timeout        int
}

type DeployConfig struct {
	id          string // role_host_[name/sequenceNum]
	role        string // etcd/mds/metaserevr
	host        string
	name        string
	sequenceNum int

	config        map[string]string
	serviceConfig map[string]string
	variables     *Variables
	sshConfig     SshConfig
}

type FilterOption struct {
	Id   string
	Role string
	Host string
}

func formatId(role, host, name string) string {
	return fmt.Sprintf("%s_%s_%s", role, host, name)
}

func newVariables(m map[string]interface{}) (*Variables, error) {
	vars := NewVariables()
	if m == nil || len(m) == 0 {
		return vars, nil
	}

	for k, v := range m {
		value := ""
		if utils.IsString(v) {
			value = v.(string)
		} else if utils.IsInt(v) {
			value = strconv.Itoa(v.(int))
		} else if !utils.IsString(v) {
			return nil, fmt.Errorf("unsupport value type for variable '%s'", k)
		}
		vars.Register(Variable{k, "", value, false})
	}

	return vars, nil
}

func NewDeployConfig(role, host, name string, sequenceNum int, config map[string]interface{}) (*DeployConfig, error) {
	var vars *Variables
	var err error
	c := map[string]string{}
	for k, v := range config {
		if k == KEY_VARIABLE && !utils.IsString(v) {
			vars, err = newVariables(v.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
		} else if utils.IsString(v) {
			c[k] = v.(string)
		} else if utils.IsInt(v) {
			c[k] = strconv.Itoa(v.(int))
		} else {
			return nil, fmt.Errorf("unsupport value type for config key '%s'", k)
		}
	}

	if name == "" {
		name = strconv.Itoa(sequenceNum)
	}
	dc := &DeployConfig{
		id:            formatId(role, host, name),
		role:          role,
		host:          host,
		name:          name,
		sequenceNum:   sequenceNum,
		config:        c,
		serviceConfig: map[string]string{},
		variables:     vars,
	}
	return dc, nil
}

func (dc *DeployConfig) check() error {
	if host := dc.GetHost(); len(host) == 0 {
		return fmt.Errorf("'%s' is an invalid host", host)
	} else if name := dc.GetName(); len(name) == 0 {
		return fmt.Errorf("'%s' is an invalid name", name)
	}

	for _, key := range []string{KEY_USER, KEY_PRIVATE_KEY_FILE, KEY_CONTAINER_IMAGE} {
		if !utils.IsString(dc.GetConfig(key)) {
			return fmt.Errorf("'%s' must be a string", key)
		} else if len(dc.GetConfig(key)) == 0 {
			return fmt.Errorf("'%s' is an invalid string", key)
		}
	}

	for _, key := range []string{KEY_SSH_PORT} {
		value := dc.GetConfig(key)
		n, err := strconv.Atoi(value)
		if err != nil || n <= 0 {
			return fmt.Errorf("'%s(%s)' is an invalid integer", key, value)
		}
	}

	return nil
}

func (dc *DeployConfig) Build() error {
	vars := dc.GetVariables()
	err := vars.Build()
	if err != nil {
		log.Error("BuildVariables", log.Field("error", err))
		return err
	}

	// step1: render variables
	if dc.host, err = vars.Rendering(dc.host); err != nil {
		return err
	} else if dc.name, err = vars.Rendering(dc.name); err != nil {
		return err
	} else if dc.id, err = vars.Rendering(dc.id); err != nil {
		return err
	}

	for k, v := range dc.config {
		v, err = vars.Rendering(v)
		if err != nil {
			return err
		}

		dc.config[k] = v
		if !excludeServiceConfig[k] {
			dc.serviceConfig[k] = v
		}
	}

	// step2: check
	err = dc.check()
	if err != nil {
		return err
	}

	// step3: generate ssh config
	sshport, err := strconv.Atoi(dc.GetConfig(KEY_SSH_PORT))
	if err != nil {
		return err
	}
	dc.sshConfig = SshConfig{
		User:           dc.GetConfig(KEY_USER),
		Host:           dc.GetHost(),
		Port:           (uint)(sshport),
		PrivateKeyPath: dc.GetConfig(KEY_PRIVATE_KEY_FILE),
		Timeout:        10, // TODO(@Wine93): get from curveadm.cfg
	}

	return nil
}

func defaultListenPort(role string) int {
	switch role {
	case ROLE_ETCD:
		return DEFAULT_ETCD_LISTEN_PEER_PORT
	case ROLE_MDS:
		return DEFAULT_MDS_LISTEN_PORT
	case ROLE_METASERVER:
		return DEFAULT_METASERVER_LISTN_PORT
	}
	return 0
}

func (dc *DeployConfig) defaultValue(key string) string {
	switch key {
	case KEY_SSH_PORT:
		return strconv.Itoa(DEFAULT_SSH_PORT)
	case KEY_PRIVATE_KEY_FILE:
		return fmt.Sprintf("/home/%s/.ssh/id_rsa", dc.GetConfig(KEY_USER))
	case KEY_LISTEN_IP:
		return dc.GetHost()
	case KEY_LISTEN_PORT:
		return strconv.Itoa(defaultListenPort(dc.GetRole()))
	case KEY_LISTEN_CLIENT_PORT:
		return strconv.Itoa(DEFAULT_ETCD_LISTEN_CLIENT_PORT)
	case KEY_LISTEN_DUMMY_PORT:
		return strconv.Itoa(DEFAULT_MDS_LISTEN_DUMMY_PORT)
	case KEY_LISTEN_EXTERNAL_IP:
		return dc.GetHost()
	}

	return ""
}

func (dc *DeployConfig) GetConfig(key string) string {
	if v, ok := dc.config[key]; ok {
		return v
	}

	return dc.defaultValue(key)
}

func (dc *DeployConfig) GetServiceConfig() map[string]string {
	return dc.serviceConfig
}

func (dc *DeployConfig) GetVariables() *Variables {
	return dc.variables
}

func (dc *DeployConfig) GetSshConfig() SshConfig {
	return dc.sshConfig
}

// wrapper interface (Get*)
func (dc *DeployConfig) GetId() string {
	return dc.id
}

func (dc *DeployConfig) GetRole() string {
	return dc.role
}

func (dc *DeployConfig) GetHost() string {
	return dc.host
}

func (dc *DeployConfig) GetName() string {
	return dc.name
}

func (dc *DeployConfig) GetSequence() int {
	return dc.sequenceNum
}

func (dc *DeployConfig) GetUser() string {
	return dc.GetConfig(KEY_USER)
}

func (dc *DeployConfig) GetSshPort() int {
	sshPort, _ := strconv.Atoi(dc.GetConfig(KEY_SSH_PORT))
	return sshPort
}

func (dc *DeployConfig) GetPrivateKeyFile() string {
	return dc.GetConfig(KEY_PRIVATE_KEY_FILE)
}

func (dc *DeployConfig) GetContainerImage() string {
	return dc.GetConfig(KEY_CONTAINER_IMAGE)
}

func (dc *DeployConfig) GetLogDir() string {
	return dc.config[KEY_LOG_DIR]
}

func (dc *DeployConfig) GetDataDir() string {
	return dc.config[KEY_DATA_DIR]
}

// /usr/local/curvefs
func (dc *DeployConfig) GetCurveFSPrefix() string {
	return DEFAULT_CURVEFS_DIR
}

// /usr/local/curvefs/mds
func (dc *DeployConfig) GetServicePrefix() string {
	return fmt.Sprintf("%s/%s", DEFAULT_CURVEFS_DIR, dc.GetRole())
}

// /usr/local/curvefs/mds/sbin
func (dc *DeployConfig) GetServiceSbinDir() string {
	return dc.GetServicePrefix() + "/sbin"
}

func FilterDeployConfig(deployConfigs []*DeployConfig, options FilterOption) []*DeployConfig {
	dcs := []*DeployConfig{}
	for _, dc := range deployConfigs {
		id := dc.GetId()
		role := dc.GetRole()
		host := dc.GetHost()
		if (options.Id == "*" || ExtractDcId(options.Id) == id) &&
			(options.Role == "*" || options.Role == role) &&
			(options.Host == "*" || options.Host == host) {
			dcs = append(dcs, dc)
		}
	}

	return dcs
}
