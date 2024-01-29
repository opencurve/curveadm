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

// __SIGN_BY_WINE93__

package topology

import (
	"fmt"
	"path"

	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/variable"
)

const (
	// service project layout
	LAYOUT_CURVEFS_ROOT_DIR                  = "/curvefs"
	LAYOUT_CURVEBS_ROOT_DIR                  = "/curvebs"
	LAYOUT_PLAYGROUND_ROOT_DIR               = "playground"
	LAYOUT_CONF_SRC_DIR                      = "/conf"
	LAYOUT_SERVICE_BIN_DIR                   = "/sbin"
	LAYOUT_SERVICE_CONF_DIR                  = "/conf"
	LAYOUT_SERVICE_LOG_DIR                   = "/logs"
	LAYOUT_SERVICE_DATA_DIR                  = "/data"
	LAYOUT_TOOLS_DIR                         = "/tools"
	LAYOUT_TOOLS_V2_DIR                      = "/tools-v2"
	LAYOUT_CURVEBS_CHUNKFILE_POOL_DIR        = "chunkfilepool"
	LAYOUT_CURVEBS_COPYSETS_DIR              = "copysets"
	LAYOUT_CURVEBS_RECYCLER_DIR              = "recycler"
	LAYOUT_CURVEBS_TOOLS_CONFIG_SYSTEM_PATH  = "/etc/curve/tools.conf"
	LAYOUT_CURVEFS_TOOLS_CONFIG_SYSTEM_PATH  = "/etc/curvefs/tools.conf"
	LAYOUT_CURVE_TOOLS_V2_CONFIG_SYSTEM_PATH = "/etc/curve/curve.yaml"
	LAYOUT_CORE_SYSTEM_DIR                   = "/core"

	BINARY_CURVEBS_TOOL     = "curvebs-tool"
	BINARY_CURVEBS_FORMAT   = "curve_format"
	BINARY_CURVEFS_TOOL     = "curvefs_tool"
	BINARY_CURVE_TOOL_V2    = "curve"
	METAFILE_CHUNKFILE_POOL = "chunkfilepool.meta"
	METAFILE_CHUNKSERVER_ID = "chunkserver.dat"
)

var (
	DefaultCurveBSDeployConfig = &DeployConfig{kind: KIND_CURVEBS}
	DefaultCurveFSDeployConfig = &DeployConfig{kind: KIND_CURVEFS}

	ServiceConfigs = map[string][]string{
		ROLE_ETCD:          []string{"etcd.conf"},
		ROLE_MDS:           []string{"mds.conf"},
		ROLE_CHUNKSERVER:   []string{"chunkserver.conf", "cs_client.conf", "s3.conf"},
		ROLE_SNAPSHOTCLONE: []string{"snapshotclone.conf", "snap_client.conf", "s3.conf", "nginx.conf"},
		ROLE_METASERVER:    []string{"metaserver.conf"},
	}
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
func (dc *DeployConfig) GetHostname() string                 { return dc.hostname }
func (dc *DeployConfig) GetName() string                     { return dc.name }
func (dc *DeployConfig) GetInstances() int                   { return dc.instances }
func (dc *DeployConfig) GetHostSequence() int                { return dc.hostSequence }
func (dc *DeployConfig) GetInstancesSequence() int           { return dc.instancesSequence }
func (dc *DeployConfig) GetServiceConfig() map[string]string { return dc.serviceConfig }
func (dc *DeployConfig) GetVariables() *variable.Variables   { return dc.variables }

// (2): config item
func (dc *DeployConfig) GetPrefix() string           { return dc.getString(CONFIG_PREFIX) }
func (dc *DeployConfig) GetReportUsage() bool        { return dc.getBool(CONFIG_REPORT_USAGE) }
func (dc *DeployConfig) GetContainerImage() string   { return dc.getString(CONFIG_CONTAINER_IMAGE) }
func (dc *DeployConfig) GetLogDir() string           { return dc.getString(CONFIG_LOG_DIR) }
func (dc *DeployConfig) GetDataDir() string          { return dc.getString(CONFIG_DATA_DIR) }
func (dc *DeployConfig) GetCoreDir() string          { return dc.getString(CONFIG_CORE_DIR) }
func (dc *DeployConfig) GetEnv() string              { return dc.getString(CONFIG_ENV) }
func (dc *DeployConfig) GetListenIp() string         { return dc.getString(CONFIG_LISTEN_IP) }
func (dc *DeployConfig) GetListenPort() int          { return dc.getInt(CONFIG_LISTEN_PORT) }
func (dc *DeployConfig) GetListenClientPort() int    { return dc.getInt(CONFIG_LISTEN_CLIENT_PORT) }
func (dc *DeployConfig) GetListenDummyPort() int     { return dc.getInt(CONFIG_LISTEN_DUMMY_PORT) }
func (dc *DeployConfig) GetListenProxyPort() int     { return dc.getInt(CONFIG_LISTEN_PROXY_PORT) }
func (dc *DeployConfig) GetListenExternalIp() string { return dc.getString(CONFIG_LISTEN_EXTERNAL_IP) }
func (dc *DeployConfig) GetCopysets() int            { return dc.getInt(CONFIG_COPYSETS) }
func (dc *DeployConfig) GetS3AccessKey() string      { return dc.getString(CONFIG_S3_ACCESS_KEY) }
func (dc *DeployConfig) GetS3SecretKey() string      { return dc.getString(CONFIG_S3_SECRET_KEY) }
func (dc *DeployConfig) GetS3Address() string        { return dc.getString(CONFIG_S3_ADDRESS) }
func (dc *DeployConfig) GetS3BucketName() string     { return dc.getString(CONFIG_S3_BUCKET_NAME) }
func (dc *DeployConfig) GetEnableRDMA() bool         { return dc.getBool(CONFIG_ENABLE_RDMA) }
func (dc *DeployConfig) GetEnableRenameAt2() bool    { return dc.getBool(CONFIG_ENABLE_RENAMEAT2) }
func (dc *DeployConfig) GetEtcdAuthEnable() bool     { return dc.getBool(CONFIG_ETCD_AUTH_ENABLE) }
func (dc *DeployConfig) GetEtcdAuthUsername() string { return dc.getString(CONFIG_ETCD_AUTH_USERNAME) }
func (dc *DeployConfig) GetEtcdAuthPassword() string { return dc.getString(CONFIG_ETCD_AUTH_PASSWORD) }
func (dc *DeployConfig) GetEnableChunkfilePool() bool {
	return dc.getBool(CONFIG_ENABLE_CHUNKFILE_POOL)
}

func (dc *DeployConfig) GetEnableExternalServer() bool {
	return dc.getBool(CONFIG_ENABLE_EXTERNAL_SERVER)
}

func (dc *DeployConfig) GetListenExternalPort() int {
	if dc.GetEnableExternalServer() {
		return dc.getInt(CONFIG_LISTEN_EXTERNAL_PORT)
	}
	return dc.GetListenPort()
}

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
 *  ├── snapshotclone
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
type (
	ConfFile struct {
		Name       string
		Path       string
		SourcePath string
	}

	Layout struct {
		// project: curvebs/curvefs
		ProjectRootDir string // /curvebs

		PlaygroundRootDir string // /curvebs/playground

		// service
		ServiceRootDir     string // /curvebs/mds
		ServiceBinDir      string // /curvebs/mds/sbin
		ServiceConfDir     string // /curvebs/mds/conf
		ServiceLogDir      string // /curvebs/mds/logs
		ServiceDataDir     string // /curvebs/mds/data
		ServiceConfPath    string // /curvebs/mds/conf/mds.conf
		ServiceConfSrcPath string // /curvebs/conf/mds.conf
		ServiceConfFiles   []ConfFile

		// tools
		ToolsRootDir        string // /curvebs/tools
		ToolsBinDir         string // /curvebs/tools/sbin
		ToolsDataDir        string // /curvebs/tools/data
		ToolsConfDir        string // /curvebs/tools/conf
		ToolsConfPath       string // /curvebs/tools/conf/tools.conf
		ToolsConfSrcPath    string // /curvebs/conf/tools.conf
		ToolsConfSystemPath string // /etc/curve/tools.conf
		ToolsBinaryPath     string // /curvebs/tools/sbin/curvebs-tool

		// tools-v2
		ToolsV2ConfSrcPath    string // /curvebs/conf/curve.yaml
		ToolsV2ConfSystemPath string // /etc/curve/curve.yaml
		ToolsV2BinaryPath     string // /curvebs/tools-v2/sbin/curve

		// format
		FormatBinaryPath      string // /curvebs/tools/sbin/curve_format
		ChunkfilePoolRootDir  string // /curvebs/chunkserver/data
		ChunkfilePoolDir      string // /curvebs/chunkserver/data/chunkfilepool
		ChunkfilePoolMetaPath string // /curvebs/chunkserver/data/chunkfilepool.meta

		// core
		CoreSystemDir string
	}
)

func (dc *DeployConfig) GetProjectLayout() Layout {
	kind := dc.GetKind()
	role := dc.GetRole()
	// project
	root := utils.Choose(kind == KIND_CURVEBS, LAYOUT_CURVEBS_ROOT_DIR, LAYOUT_CURVEFS_ROOT_DIR)

	// service
	confSrcDir := root + LAYOUT_CONF_SRC_DIR
	serviceRootDir := dc.GetPrefix()
	serviceConfDir := fmt.Sprintf("%s/conf", serviceRootDir)
	serviceConfFiles := []ConfFile{}
	for _, item := range ServiceConfigs[role] {
		serviceConfFiles = append(serviceConfFiles, ConfFile{
			Name:       item,
			Path:       fmt.Sprintf("%s/%s", serviceConfDir, item),
			SourcePath: fmt.Sprintf("%s/%s", confSrcDir, item),
		})
	}

	// tools
	toolsRootDir := root + LAYOUT_TOOLS_DIR
	toolsBinDir := toolsRootDir + LAYOUT_SERVICE_BIN_DIR
	toolsConfDir := toolsRootDir + LAYOUT_SERVICE_CONF_DIR
	toolsBinaryName := utils.Choose(kind == KIND_CURVEBS, BINARY_CURVEBS_TOOL, BINARY_CURVEFS_TOOL)
	toolsConfSystemPath := utils.Choose(kind == KIND_CURVEBS,
		LAYOUT_CURVEBS_TOOLS_CONFIG_SYSTEM_PATH,
		LAYOUT_CURVEFS_TOOLS_CONFIG_SYSTEM_PATH)

	// tools-v2
	toolsV2RootDir := root + LAYOUT_TOOLS_V2_DIR
	toolsV2BinDir := toolsV2RootDir + LAYOUT_SERVICE_BIN_DIR
	toolsV2BinaryName := BINARY_CURVE_TOOL_V2
	toolsV2ConfSystemPath := LAYOUT_CURVE_TOOLS_V2_CONFIG_SYSTEM_PATH

	// format
	chunkserverDataDir := fmt.Sprintf("%s/%s%s", root, ROLE_CHUNKSERVER, LAYOUT_SERVICE_DATA_DIR)

	return Layout{
		// project
		ProjectRootDir: root,

		// playground
		PlaygroundRootDir: path.Join(root, LAYOUT_PLAYGROUND_ROOT_DIR),

		// service
		ServiceRootDir:     serviceRootDir,
		ServiceBinDir:      serviceRootDir + LAYOUT_SERVICE_BIN_DIR,
		ServiceConfDir:     serviceRootDir + LAYOUT_SERVICE_CONF_DIR,
		ServiceLogDir:      serviceRootDir + LAYOUT_SERVICE_LOG_DIR,
		ServiceDataDir:     serviceRootDir + LAYOUT_SERVICE_DATA_DIR,
		ServiceConfPath:    fmt.Sprintf("%s/%s.conf", serviceConfDir, role),
		ServiceConfSrcPath: fmt.Sprintf("%s/%s.conf", confSrcDir, role),
		ServiceConfFiles:   serviceConfFiles,

		// tools
		ToolsRootDir:        toolsRootDir,
		ToolsBinDir:         toolsRootDir + LAYOUT_SERVICE_BIN_DIR,
		ToolsDataDir:        toolsRootDir + LAYOUT_SERVICE_DATA_DIR,
		ToolsConfDir:        toolsRootDir + LAYOUT_SERVICE_CONF_DIR,
		ToolsConfPath:       fmt.Sprintf("%s/tools.conf", toolsConfDir),
		ToolsConfSrcPath:    fmt.Sprintf("%s/tools.conf", confSrcDir),
		ToolsConfSystemPath: toolsConfSystemPath,
		ToolsBinaryPath:     fmt.Sprintf("%s/%s", toolsBinDir, toolsBinaryName),

		// toolsv2
		ToolsV2ConfSrcPath:    fmt.Sprintf("%s/curve.yaml", confSrcDir),
		ToolsV2ConfSystemPath: toolsV2ConfSystemPath,
		ToolsV2BinaryPath:     fmt.Sprintf("%s/%s", toolsV2BinDir, toolsV2BinaryName),

		// format
		FormatBinaryPath:      fmt.Sprintf("%s/%s", toolsBinDir, BINARY_CURVEBS_FORMAT),
		ChunkfilePoolRootDir:  chunkserverDataDir,
		ChunkfilePoolDir:      fmt.Sprintf("%s/%s", chunkserverDataDir, LAYOUT_CURVEBS_CHUNKFILE_POOL_DIR),
		ChunkfilePoolMetaPath: fmt.Sprintf("%s/%s", chunkserverDataDir, METAFILE_CHUNKFILE_POOL),

		// core
		CoreSystemDir: LAYOUT_CORE_SYSTEM_DIR,
	}
}

func GetProjectLayout(kind, role string) Layout {
	dc := DeployConfig{kind: kind, role: role}
	return dc.GetProjectLayout()
}

func GetCurveBSProjectLayout() Layout {
	return DefaultCurveBSDeployConfig.GetProjectLayout()
}

func GetCurveFSProjectLayout() Layout {
	return DefaultCurveFSDeployConfig.GetProjectLayout()
}
