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

package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
	configure "github.com/opencurve/curveadm/internal/configure/curveadm"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/plugin"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/log"
)

type CurveAdm struct {
	// project layout
	rootDir       string
	dataDir       string
	pluginDir     string
	logDir        string
	tempDir       string
	inventoryPath string
	logpath       string
	config        *configure.CurveAdmConfig

	// data pipeline
	in            io.Reader
	out           io.Writer
	err           io.Writer
	storage       *storage.Storage
	memStorage    *utils.SafeMap
	pluginManager *plugin.PluginManager

	clusterId           int    // current cluster id
	clusterUUId         string // current cluster uuid
	clusterName         string // current cluster name
	clusterTopologyData string // cluster topology
}

/*
 * $HOME/.curveadm
 *   - curveadm.cfg
 *   - server.yaml
 *   - /bin/curveadm
 *   - /data/curveadm.db
 *   - /plugins/{shell,file,polarfs}
 *   - /logs/2006-01-02_15-04-05.log
 *   - /temp/
 */
func NewCurveAdm() (*CurveAdm, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	rootDir := fmt.Sprintf("%s/.curveadm", home)
	curveadm := &CurveAdm{
		rootDir:       rootDir,
		dataDir:       rootDir + "/data",
		pluginDir:     rootDir + "/plugins",
		logDir:        rootDir + "/logs",
		tempDir:       rootDir + "/temp",
		inventoryPath: rootDir + "/server.yaml",
	}

	err = curveadm.init()
	return curveadm, err
}

func (curveadm *CurveAdm) init() error {
	// create directory
	dirs := []string{curveadm.rootDir, curveadm.dataDir, curveadm.pluginDir, curveadm.logDir, curveadm.tempDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	// parse curveadm.cfg
	confpath := fmt.Sprintf("%s/curveadm.cfg", curveadm.rootDir)
	config, err := configure.ParseCurveAdmConfig(confpath)
	if err != nil {
		return err
	}

	// init logger
	now := time.Now().Format("2006-01-02_15-04-05")
	logpath := fmt.Sprintf("%s/curveadm-%s.log", curveadm.logDir, now)
	if err := log.Init(config.LogLevel, logpath); err != nil {
		return err
	} else {
		log.Info("InitLogger",
			log.Field("logPath", logpath),
			log.Field("logLevel", config.LogLevel))
	}

	// new storage: create table in sqlite
	dbpath := fmt.Sprintf("%s/curveadm.db", curveadm.dataDir)
	s, err := storage.NewStorage(dbpath)
	if err != nil {
		log.Error("NewStorage", log.Field("error", err))
		return err
	}

	// get current cluster
	cluster, err := s.GetCurrentCluster()
	if err != nil {
		log.Error("GetCurrentCluster", log.Field("error", err))
		return err
	} else {
		log.Info("GetCurrentCluster",
			log.Field("clusterId", cluster.Id),
			log.Field("clusterName", cluster.Name))
	}

	curveadm.logpath = logpath
	curveadm.config = config
	curveadm.in = os.Stdin
	curveadm.out = os.Stdout
	curveadm.err = os.Stderr
	curveadm.storage = s
	curveadm.memStorage = utils.NewSafeMap()
	curveadm.pluginManager = plugin.NewPluginManager(curveadm.pluginDir)
	curveadm.clusterId = cluster.Id
	curveadm.clusterUUId = cluster.UUId
	curveadm.clusterName = cluster.Name
	curveadm.clusterTopologyData = cluster.Topology

	return nil
}

func (curveadm *CurveAdm) RootDir() string                      { return curveadm.rootDir }
func (curveadm *CurveAdm) DataDir() string                      { return curveadm.dataDir }
func (curveadm *CurveAdm) PluginDir() string                    { return curveadm.pluginDir }
func (curveadm *CurveAdm) LogDir() string                       { return curveadm.logDir }
func (curveadm *CurveAdm) TempDir() string                      { return curveadm.tempDir }
func (curveadm *CurveAdm) InventoryPath() string                { return curveadm.inventoryPath }
func (curveadm *CurveAdm) LogPath() string                      { return curveadm.logpath }
func (curveadm *CurveAdm) Config() *configure.CurveAdmConfig    { return curveadm.config }
func (curveadm *CurveAdm) In() io.Reader                        { return curveadm.in }
func (curveadm *CurveAdm) Out() io.Writer                       { return curveadm.out }
func (curveadm *CurveAdm) Err() io.Writer                       { return curveadm.err }
func (curveadm *CurveAdm) Storage() *storage.Storage            { return curveadm.storage }
func (curveadm *CurveAdm) MemStorage() *utils.SafeMap           { return curveadm.memStorage }
func (curveadm *CurveAdm) PluginManager() *plugin.PluginManager { return curveadm.pluginManager }
func (curveadm *CurveAdm) ClusterId() int                       { return curveadm.clusterId }
func (curveadm *CurveAdm) ClusterUUId() string                  { return curveadm.clusterUUId }
func (curveadm *CurveAdm) ClusterName() string                  { return curveadm.clusterName }
func (curveadm *CurveAdm) ClusterTopologyData() string          { return curveadm.clusterTopologyData }

func (curveadm *CurveAdm) ParseTopology() ([]*topology.DeployConfig, error) {
	return topology.ParseTopology(curveadm.clusterTopologyData)
}

func (curveadm *CurveAdm) GetServiceId(dcId string) string {
	serviceId := fmt.Sprintf("%s_%s", curveadm.ClusterUUId(), dcId)
	return utils.MD5Sum(serviceId)[:12]
}

func (curveadm *CurveAdm) FilterDeployConfig(deployConfigs []*topology.DeployConfig,
	options topology.FilterOption) []*topology.DeployConfig {
	dcs := []*topology.DeployConfig{}
	for _, dc := range deployConfigs {
		dcId := dc.GetId()
		role := dc.GetRole()
		host := dc.GetHost()
		serviceId := curveadm.GetServiceId(dcId)
		if (options.Id == "*" || options.Id == serviceId) &&
			(options.Role == "*" || options.Role == role) &&
			(options.Host == "*" || options.Host == host) {
			dcs = append(dcs, dc)
		}
	}

	return dcs
}

func (curveadm *CurveAdm) SudoAlias() string { return curveadm.config.GetSudoAlias() }
func (curveadm *CurveAdm) SSHTimeout() int   { return curveadm.config.GetSSHTimeout() }

func (curveadm *CurveAdm) WriteOut(format string, a ...interface{}) (int, error) {
	output := fmt.Sprintf(format, a...)
	return curveadm.out.Write([]byte(output))
}

func (curveadm *CurveAdm) NewPromptError(err error, prompt string) utils.PromptError {
	if prompt == "" {
		prompt = color.CyanString("See log file for detail: %s", curveadm.LogPath())
	}
	return utils.PromptError{err, prompt}
}
