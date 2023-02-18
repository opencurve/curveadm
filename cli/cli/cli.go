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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	comm "github.com/opencurve/curveadm/internal/common"
	configure "github.com/opencurve/curveadm/internal/configure/curveadm"
	"github.com/opencurve/curveadm/internal/configure/hosts"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/storage"
	tools "github.com/opencurve/curveadm/internal/tools/upgrade"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	log "github.com/opencurve/curveadm/pkg/log/glg"
	"github.com/opencurve/curveadm/pkg/module"
)

type CurveAdm struct {
	// project layout
	rootDir   string
	dataDir   string
	pluginDir string
	logDir    string
	tempDir   string
	dbpath    string
	logpath   string
	config    *configure.CurveAdmConfig

	// data pipeline
	in         io.Reader
	out        io.Writer
	err        io.Writer
	storage    *storage.Storage
	memStorage *utils.SafeMap

	// properties (hosts/cluster)
	hosts               string         // hosts
	disks               string         // disks of yaml data
	diskRecords         []storage.Disk // disk records stored in database
	clusterId           int            // current cluster id
	clusterUUId         string         // current cluster uuid
	clusterName         string         // current cluster name
	clusterTopologyData string         // cluster topology
	clusterPoolData     string         // cluster pool
}

/*
 * $HOME/.curveadm
 *   - curveadm.cfg
 *   - /bin/curveadm
 *   - /data/curveadm.db
 *   - /plugins/{shell,file,polarfs}
 *   - /logs/2006-01-02_15-04-05.log
 *   - /temp/
 */
func NewCurveAdm() (*CurveAdm, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errno.ERR_GET_USER_HOME_DIR_FAILED.E(err)
	}

	rootDir := fmt.Sprintf("%s/.curveadm", home)
	curveadm := &CurveAdm{
		rootDir:   rootDir,
		dataDir:   path.Join(rootDir, "data"),
		pluginDir: path.Join(rootDir, "plugins"),
		logDir:    path.Join(rootDir, "logs"),
		tempDir:   path.Join(rootDir, "temp"),
	}

	err = curveadm.init()
	if err != nil {
		return nil, err
	}

	go curveadm.detectVersion()
	return curveadm, nil
}

func (curveadm *CurveAdm) init() error {
	// (1) Create directory
	dirs := []string{
		curveadm.rootDir,
		curveadm.dataDir,
		curveadm.pluginDir,
		curveadm.logDir,
		curveadm.tempDir,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return errno.ERR_CREATE_CURVEADM_SUBDIRECTORY_FAILED.E(err)
		}
	}

	// (2) Parse curveadm.cfg
	confpath := fmt.Sprintf("%s/curveadm.cfg", curveadm.rootDir)
	config, err := configure.ParseCurveAdmConfig(confpath)
	if err != nil {
		return err
	}
	configure.ReplaceGlobals(config)

	// (3) Init logger
	now := time.Now().Format("2006-01-02_15-04-05")
	logpath := fmt.Sprintf("%s/curveadm-%s.log", curveadm.logDir, now)
	if err := log.Init(config.GetLogLevel(), logpath); err != nil {
		return errno.ERR_INIT_LOGGER_FAILED.E(err)
	} else {
		log.Info("Init logger success",
			log.Field("LogPath", logpath),
			log.Field("LogLevel", config.GetLogLevel()))
	}

	// (4) Init error code
	errno.Init(logpath)

	// (5) New storage: create table in sqlite
	dbpath := fmt.Sprintf("%s/curveadm.db", curveadm.dataDir)
	s, err := storage.NewStorage(dbpath)
	if err != nil {
		log.Error("Init SQLite database failed",
			log.Field("Error", err))
		return errno.ERR_INIT_SQL_DATABASE_FAILED.E(err)
	}

	// (6) Get hosts
	var hosts storage.Hosts
	hostses, err := s.GetHostses()
	if err != nil {
		log.Error("Get hosts failed",
			log.Field("Error", err))
		return errno.ERR_GET_HOSTS_FAILED.E(err)
	} else if len(hostses) == 1 {
		hosts = hostses[0]
	}

	// (7) Get current cluster
	cluster, err := s.GetCurrentCluster()
	if err != nil {
		log.Error("Get current cluster failed",
			log.Field("Error", err))
		return errno.ERR_GET_CURRENT_CLUSTER_FAILED.E(err)
	} else {
		log.Info("Get current cluster success",
			log.Field("ClusterId", cluster.Id),
			log.Field("ClusterName", cluster.Name))
	}

	// (8) Get Disks
	var disks storage.Disks
	diskses, err := s.GetDisks()
	if err != nil {
		log.Error("Get disks failed", log.Field("Error", err))
		return errno.ERR_GET_DISKS_FAILED.E(err)
	} else if len(diskses) > 0 {
		disks = diskses[0]
	}

	// (9) Get Disk Records
	diskRecords, err := s.GetDisk("all")
	if err != nil {
		log.Error("Get disk records failed", log.Field("Error", err))
		return errno.ERR_GET_DISK_RECORDS_FAILED.E(err)
	}

	curveadm.dbpath = dbpath
	curveadm.logpath = logpath
	curveadm.config = config
	curveadm.in = os.Stdin
	curveadm.out = os.Stdout
	curveadm.err = os.Stderr
	curveadm.storage = s
	curveadm.memStorage = utils.NewSafeMap()
	curveadm.hosts = hosts.Data
	curveadm.disks = disks.Data
	curveadm.diskRecords = diskRecords
	curveadm.clusterId = cluster.Id
	curveadm.clusterUUId = cluster.UUId
	curveadm.clusterName = cluster.Name
	curveadm.clusterTopologyData = cluster.Topology
	curveadm.clusterPoolData = cluster.Pool

	return nil
}

func (curveadm *CurveAdm) detectVersion() {
	latestVersion, err := tools.GetLatestVersion(Version)
	if err != nil || len(latestVersion) == 0 {
		return
	}

	versions, err := curveadm.Storage().GetVersions()
	if err != nil {
		return
	} else if len(versions) > 0 {
		pendingVersion := versions[0].Version
		if pendingVersion == latestVersion {
			return
		}
	}

	curveadm.Storage().SetVersion(latestVersion, "")
}

func (curveadm *CurveAdm) Upgrade() (bool, error) {
	if curveadm.config.GetAutoUpgrade() == false {
		return false, nil
	}

	versions, err := curveadm.Storage().GetVersions()
	if err != nil || len(versions) == 0 {
		return false, nil
	}

	// (1) skip upgrade if the pending version is stale
	latestVersion := versions[0].Version
	err, yes := tools.IsLatest(Version, strings.TrimPrefix(latestVersion, "v"))
	if err != nil || yes {
		return false, nil
	}

	// (2) skip upgrade if user has confirmed
	day := time.Now().Format("2006-01-02")
	lastConfirm := versions[0].LastConfirm
	if day == lastConfirm {
		return false, nil
	}

	curveadm.Storage().SetVersion(latestVersion, day)
	pass := tui.ConfirmYes(tui.PromptAutoUpgrade(latestVersion))
	if !pass {
		return false, errno.ERR_CANCEL_OPERATION
	}

	err = tools.Upgrade(latestVersion)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (curveadm *CurveAdm) RootDir() string                   { return curveadm.rootDir }
func (curveadm *CurveAdm) DataDir() string                   { return curveadm.dataDir }
func (curveadm *CurveAdm) PluginDir() string                 { return curveadm.pluginDir }
func (curveadm *CurveAdm) LogDir() string                    { return curveadm.logDir }
func (curveadm *CurveAdm) TempDir() string                   { return curveadm.tempDir }
func (curveadm *CurveAdm) DBPath() string                    { return curveadm.dbpath }
func (curveadm *CurveAdm) LogPath() string                   { return curveadm.logpath }
func (curveadm *CurveAdm) Config() *configure.CurveAdmConfig { return curveadm.config }
func (curveadm *CurveAdm) SudoAlias() string                 { return curveadm.config.GetSudoAlias() }
func (curveadm *CurveAdm) SSHTimeout() int                   { return curveadm.config.GetSSHTimeout() }
func (curveadm *CurveAdm) In() io.Reader                     { return curveadm.in }
func (curveadm *CurveAdm) Out() io.Writer                    { return curveadm.out }
func (curveadm *CurveAdm) Err() io.Writer                    { return curveadm.err }
func (curveadm *CurveAdm) Storage() *storage.Storage         { return curveadm.storage }
func (curveadm *CurveAdm) MemStorage() *utils.SafeMap        { return curveadm.memStorage }
func (curveadm *CurveAdm) Hosts() string                     { return curveadm.hosts }
func (curveadm *CurveAdm) Disks() string                     { return curveadm.disks }
func (curveadm *CurveAdm) DiskRecords() []storage.Disk       { return curveadm.diskRecords }
func (curveadm *CurveAdm) ClusterId() int                    { return curveadm.clusterId }
func (curveadm *CurveAdm) ClusterUUId() string               { return curveadm.clusterUUId }
func (curveadm *CurveAdm) ClusterName() string               { return curveadm.clusterName }
func (curveadm *CurveAdm) ClusterTopologyData() string       { return curveadm.clusterTopologyData }
func (curveadm *CurveAdm) ClusterPoolData() string           { return curveadm.clusterPoolData }

func (curveadm *CurveAdm) GetHost(host string) (*hosts.HostConfig, error) {
	if len(curveadm.Hosts()) == 0 {
		return nil, errno.ERR_HOST_NOT_FOUND.
			F("host: %s", host)
	}
	hcs, err := hosts.ParseHosts(curveadm.Hosts())
	if err != nil {
		return nil, err
	}

	for _, hc := range hcs {
		if hc.GetHost() == host {
			return hc, nil
		}
	}
	return nil, errno.ERR_HOST_NOT_FOUND.
		F("host: %s", host)
}

func (curveadm *CurveAdm) ParseTopologyData(data string) ([]*topology.DeployConfig, error) {
	ctx := topology.NewContext()
	hcs, err := hosts.ParseHosts(curveadm.Hosts())
	if err != nil {
		return nil, err
	}
	for _, hc := range hcs {
		ctx.Add(hc.GetHost(), hc.GetHostname())
	}

	dcs, err := topology.ParseTopology(data, ctx)
	if err != nil {
		return nil, err
	} else if len(dcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_IN_TOPOLOGY
	}
	return dcs, err
}

func (curveadm *CurveAdm) ParseTopology() ([]*topology.DeployConfig, error) {
	if curveadm.ClusterId() == -1 {
		return nil, errno.ERR_NO_CLUSTER_SPECIFIED
	}
	return curveadm.ParseTopologyData(curveadm.ClusterTopologyData())
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

func (curveadm *CurveAdm) FilterDeployConfigByRole(dcs []*topology.DeployConfig,
	role string) []*topology.DeployConfig {
	options := topology.FilterOption{Id: "*", Role: role, Host: "*"}
	return curveadm.FilterDeployConfig(dcs, options)
}

func (curveadm *CurveAdm) GetServiceId(dcId string) string {
	serviceId := fmt.Sprintf("%s_%s", curveadm.ClusterUUId(), dcId)
	return utils.MD5Sum(serviceId)[:12]
}

func (curveadm *CurveAdm) GetContainerId(serviceId string) (string, error) {
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return "", errno.ERR_GET_SERVICE_CONTAINER_ID_FAILED
	} else if len(containerId) == 0 {
		return "", errno.ERR_SERVICE_CONTAINER_ID_NOT_FOUND
	}
	return containerId, nil
}

// FIXME
func (curveadm *CurveAdm) IsSkip(dc *topology.DeployConfig) bool {
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	return err == nil && len(containerId) == 0 && dc.GetRole() == topology.ROLE_SNAPSHOTCLONE
}

func (curveadm *CurveAdm) GetVolumeId(host, user, volume string) string {
	volumeId := fmt.Sprintf("curvebs_volume_%s_%s_%s", host, user, volume)
	return utils.MD5Sum(volumeId)[:12]
}

func (curveadm *CurveAdm) GetFilesystemId(host, mountPoint string) string {
	filesystemId := fmt.Sprintf("curvefs_filesystem_%s_%s", host, mountPoint)
	return utils.MD5Sum(filesystemId)[:12]
}

func (curveadm *CurveAdm) ExecOptions() module.ExecOptions {
	return module.ExecOptions{
		ExecWithSudo:   true,
		ExecInLocal:    false,
		ExecSudoAlias:  curveadm.config.GetSudoAlias(),
		ExecTimeoutSec: curveadm.config.GetTimeout(),
	}
}

func (curveadm *CurveAdm) CheckId(id string) error {
	services, err := curveadm.Storage().GetServices(curveadm.ClusterId())
	if err != nil {
		return err
	}
	for _, service := range services {
		if service.Id == id {
			return nil
		}
	}
	return errno.ERR_ID_NOT_FOUND.F("id: %s", id)
}

func (curveadm *CurveAdm) CheckRole(role string) error {
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	kind := dcs[0].GetKind()
	roles := topology.CURVEBS_ROLES
	if kind == topology.KIND_CURVEFS {
		roles = topology.CURVEFS_ROLES
	}
	supported := utils.Slice2Map(roles)
	if !supported[role] {
		if kind == topology.KIND_CURVEBS {
			return errno.ERR_UNSUPPORT_CURVEBS_ROLE.
				F("role: %s", role)
		}
		return errno.ERR_UNSUPPORT_CURVEFS_ROLE.
			F("role: %s", role)
	}
	return nil
}

func (curveadm *CurveAdm) CheckHost(host string) error {
	_, err := curveadm.GetHost(host)
	return err
}

// writer for cobra command error
func (curveadm *CurveAdm) Write(p []byte) (int, error) {
	// trim prefix which generate by cobra
	p = p[len(cliutil.PREFIX_COBRA_COMMAND_ERROR):]
	return curveadm.WriteOut(string(p))
}

func (curveadm *CurveAdm) WriteOut(format string, a ...interface{}) (int, error) {
	output := fmt.Sprintf(format, a...)
	return curveadm.out.Write([]byte(output))
}

func (curveadm *CurveAdm) WriteOutln(format string, a ...interface{}) (int, error) {
	output := fmt.Sprintf(format, a...) + "\n"
	return curveadm.out.Write([]byte(output))
}

func (curveadm *CurveAdm) IsSameRole(dcs []*topology.DeployConfig) bool {
	role := dcs[0].GetRole()
	for _, dc := range dcs {
		if dc.GetRole() != role {
			return false
		}
	}
	return true
}

func (curveadm *CurveAdm) DiffTopology(data1, data2 string) ([]topology.TopologyDiff, error) {
	ctx := topology.NewContext()
	hcs, err := hosts.ParseHosts(curveadm.Hosts())
	if err != nil {
		return nil, err
	}
	for _, hc := range hcs {
		ctx.Add(hc.GetHost(), hc.GetHostname())
	}

	if len(data1) == 0 {
		return nil, errno.ERR_EMPTY_CLUSTER_TOPOLOGY
	}

	dcs, err := topology.ParseTopology(data1, ctx)
	if err != nil {
		return nil, err // err is error code
	}
	if len(dcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_IN_TOPOLOGY
	}
	return topology.DiffTopology(data1, data2, ctx)
}

func (curveadm *CurveAdm) PreAudit(now time.Time, args []string) int64 {
	if len(args) == 0 {
		return -1
	} else if args[0] == "audit" || args[0] == "__complete" {
		return -1
	}

	cwd, _ := os.Getwd()
	command := fmt.Sprintf("curveadm %s", strings.Join(args, " "))
	id, err := curveadm.Storage().InsertAuditLog(
		now, cwd, command, comm.AUDIT_STATUS_ABORT)
	if err != nil {
		log.Error("Insert audit log failed",
			log.Field("Error", err))
	}

	return id
}

func (curveadm *CurveAdm) PostAudit(id int64, ec error) {
	if id < 0 {
		return
	}

	auditLogs, err := curveadm.Storage().GetAuditLog(id)
	if err != nil {
		log.Error("Get audit log failed",
			log.Field("Error", err))
		return
	} else if len(auditLogs) != 1 {
		return
	}

	auditLog := auditLogs[0]
	status := auditLog.Status
	errorCode := 0
	if ec == nil {
		status = comm.AUDIT_STATUS_SUCCESS
	} else if errors.Is(ec, errno.ERR_CANCEL_OPERATION) {
		status = comm.AUDIT_STATUS_CANCEL
	} else {
		status = comm.AUDIT_STATUS_FAIL
		if v, ok := ec.(*errno.ErrorCode); ok {
			errorCode = v.GetCode()
		}
	}

	err = curveadm.Storage().SetAuditLogStatus(id, status, errorCode)
	if err != nil {
		log.Error("Set audit log status failed",
			log.Field("Error", err))
	}
}
