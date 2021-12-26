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
 * Created Date: 2021-12-23
 * Author: Jingli Chen (Wine93)
 */

package topology

import (
	"fmt"

	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
	"github.com/opencurve/curveadm/pkg/variable"
)

const (
	// service project layout
	LAYOUT_CURVEFS_ROOT_DIR                 = "/curvefs"
	LAYOUT_CURVEBS_ROOT_DIR                 = "/curvebs"
	LAYOUT_CONF_SRC_DIR                     = "/conf"
	LAYOUT_SERVICE_BIN_DIR                  = "/sbin"
	LAYOUT_SERVICE_CONF_DIR                 = "/conf"
	LAYOUT_SERVICE_LOG_DIR                  = "/logs"
	LAYOUT_SERVICE_DATA_DIR                 = "/data"
	LAYOUT_TOOLS_DIR                        = "/tools"
	LAYOUT_CURVEFS_TOOLS_CONFIG_SYSTEM_PATH = "/etc/curvefs/tools.conf"
	LAYOUT_CURVEBS_TOOLS_CONFIG_SYSTEM_PATH = "/etc/curve/tools.conf"
	LAYOUT_CORE_SYSTEM_DIR                  = "/core"

	BINARY_CURVEBS_TOOL = "curvebs-tool"
	BINARY_CURVEFS_TOOL = "curvefs_tool"
)

func (dc *DeployConfig) get(i *item) interface{} {
	if v, ok := dc.config[i.key]; ok {
		return v
	}

	defaultValue := i.defaultValue
	if defaultValue != nil && utils.IsFunc(defaultValue) {
		return defaultValue.(func(*DeployConfig) interface{})(dc)
	}
	return defaultValue
}

func (dc *DeployConfig) getString(i *item) string {
	v := dc.get(i)
	if v == nil {
		return ""
	}
	return v.(string)
}

func (dc *DeployConfig) getInt(i *item) int {
	v := dc.get(i)
	if v == nil {
		return 0
	}
	return v.(int)
}

func (dc *DeployConfig) getBool(i *item) bool {
	v := dc.get(i)
	if v == nil {
		return false
	}
	return v.(bool)
}

// (1): config property
func (dc *DeployConfig) GetKind() string                     { return dc.kind }
func (dc *DeployConfig) GetId() string                       { return dc.id }
func (dc *DeployConfig) GetParentId() string                 { return dc.parentId }
func (dc *DeployConfig) GetRole() string                     { return dc.role }
func (dc *DeployConfig) GetHost() string                     { return dc.host }
func (dc *DeployConfig) GetName() string                     { return dc.name }
func (dc *DeployConfig) GetReplica() int                     { return dc.replica }
func (dc *DeployConfig) GetHostSequence() int                { return dc.hostSequence }
func (dc *DeployConfig) GetReplicaSequence() int             { return dc.replicaSequence }
func (dc *DeployConfig) GetServiceConfig() map[string]string { return dc.serviceConfig }
func (dc *DeployConfig) GetVariables() *variable.Variables   { return dc.variables }

func (dc *DeployConfig) GetSSHConfig() *module.SSHConfig {
	return &module.SSHConfig{
		User:           dc.GetUser(),
		Host:           dc.GetHost(),
		Port:           (uint)(dc.GetSSHPort()),
		PrivateKeyPath: dc.GetPrivateKeyFile(),
		Timeout:        DEFAULT_SSH_TIMEOUT_SECONDS,
	}
}

// (2): config item
func (dc *DeployConfig) GetUser() string             { return dc.getString(CONFIG_USER) }
func (dc *DeployConfig) GetSSHPort() int             { return dc.getInt(CONFIG_SSH_PORT) }
func (dc *DeployConfig) GetPrivateKeyFile() string   { return dc.getString(CONFIG_PRIVATE_CONFIG_FILE) }
func (dc *DeployConfig) GetReportUsage() bool        { return dc.getBool(CONFIG_REPORT_USAGE) }
func (dc *DeployConfig) GetContainerImage() string   { return dc.getString(CONFIG_CONTAINER_IMAGE) }
func (dc *DeployConfig) GetLogDir() string           { return dc.getString(CONFIG_LOG_DIR) }
func (dc *DeployConfig) GetDataDir() string          { return dc.getString(CONFIG_DATA_DIR) }
func (dc *DeployConfig) GetCoreDir() string          { return dc.getString(CONFIG_CORE_DIR) }
func (dc *DeployConfig) GetListenIp() string         { return dc.getString(CONFIG_LISTEN_IP) }
func (dc *DeployConfig) GetListenPort() int          { return dc.getInt(CONFIG_LISTEN_PORT) }
func (dc *DeployConfig) GetListenClientPort() int    { return dc.getInt(CONFIG_LISTEN_CLIENT_PORT) }
func (dc *DeployConfig) GetListenDummyPort() int     { return dc.getInt(CONFIG_LISTEN_DUMMY_PORT) }
func (dc *DeployConfig) GetListenExternalIp() string { return dc.getString(CONFIG_LISTEN_EXTERNAL_IP) }
func (dc *DeployConfig) GetCopysets() int            { return dc.getInt(CONFIG_COPYSETS) }

// (3): service project layout
/* /curvebs
 *  ├── conf
 *  │   ├── chunkserver.conf
 *  │   ├── etcd.conf
 *  │   ├── mds.conf
 *  │   └── tools.conf
 *  ├── etcd
 *  │   ├── conf
 *  │   ├── data
 *  │   ├── log
 *  │   └── sbin
 *  ├── mds
 *  │   ├── conf
 *  │   ├── data
 *  │   ├── log
 *  │   └── sbin
 *  ├── chunkserver
 *  │   ├── conf
 *  │   ├── data
 *  │   ├── log
 *  │   └── sbin
 *  └── tools
 *      ├── conf
 *      ├── data
 *      ├── log
 *      └── sbin
 */
type Layout struct {
	// project: curvebs/curvefs
	ProjectRootDir string // /curvebs

	// service
	ServiceRootDir     string // /curvebs/mds
	ServiceBinDir      string // /curvebs/mds/sbin
	ServiceConfDir     string // /curvebs/mds/conf
	ServiceLogDir      string // /curvebs/mds/logs
	ServiceDataDir     string // /curvebs/mds/data
	ServiceConfPath    string // /curvebs/mds/conf/mds.conf
	ServiceConfSrcPath string // /curvebs/conf/mds.conf

	// tools
	ToolsRootDir        string // /curvebs/tools
	ToolsBinDir         string // /curvebs/tools/sbin
	ToolsConfDir        string // /curvebs/tools/conf
	ToolsConfPath       string // /curvebs/tools/conf/tools.conf
	ToolsConfSrcPath    string // /curvebs/conf/tools.conf
	ToolsConfSystemPath string // /etc/curve/tools.conf
	ToolsBinaryPath     string // /curvebs/tools/sbin/curvebs-tool

	// core
	CoreSystemDir string
}

func (dc *DeployConfig) GetProjectLayout() Layout {
	kind := dc.GetKind()
	role := dc.GetRole()
	root := utils.Choose(kind == KIND_CURVEBS, LAYOUT_CURVEBS_ROOT_DIR, LAYOUT_CURVEFS_ROOT_DIR)
	confSrcDir := root + LAYOUT_CONF_SRC_DIR
	serviceRootDir := fmt.Sprintf("%s/%s", root, role)
	serviceConfDir := fmt.Sprintf("%s/conf", serviceRootDir)
	toolsRootDir := root + LAYOUT_TOOLS_DIR
	toolsBinDir := toolsRootDir + LAYOUT_SERVICE_BIN_DIR
	toolsConfDir := toolsRootDir + LAYOUT_SERVICE_CONF_DIR
	toolsBinaryName := utils.Choose(kind == KIND_CURVEBS, BINARY_CURVEBS_TOOL, BINARY_CURVEFS_TOOL)
	toolsConfSystemPath := utils.Choose(kind == KIND_CURVEBS,
		LAYOUT_CURVEBS_TOOLS_CONFIG_SYSTEM_PATH,
		LAYOUT_CURVEFS_TOOLS_CONFIG_SYSTEM_PATH)

	return Layout{
		// project
		ProjectRootDir: root,
		// service
		ServiceRootDir:     serviceRootDir,
		ServiceBinDir:      serviceRootDir + LAYOUT_SERVICE_BIN_DIR,
		ServiceConfDir:     serviceRootDir + LAYOUT_SERVICE_CONF_DIR,
		ServiceLogDir:      serviceRootDir + LAYOUT_SERVICE_LOG_DIR,
		ServiceDataDir:     serviceRootDir + LAYOUT_SERVICE_DATA_DIR,
		ServiceConfPath:    fmt.Sprintf("%s/%s.conf", serviceConfDir, role),
		ServiceConfSrcPath: fmt.Sprintf("%s/%s.conf", confSrcDir, role),
		// tools
		ToolsRootDir:        toolsRootDir,
		ToolsBinDir:         toolsRootDir + LAYOUT_SERVICE_BIN_DIR,
		ToolsConfDir:        toolsRootDir + LAYOUT_SERVICE_CONF_DIR,
		ToolsConfPath:       fmt.Sprintf("%s/tools.conf", toolsConfDir),
		ToolsConfSrcPath:    fmt.Sprintf("%s/tools.conf", confSrcDir),
		ToolsConfSystemPath: toolsConfSystemPath,
		ToolsBinaryPath:     fmt.Sprintf("%s/%s", toolsBinDir, toolsBinaryName),
		// core
		CoreSystemDir: LAYOUT_CORE_SYSTEM_DIR,
	}
}
