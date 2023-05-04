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
 * Created Date: 2022-07-27
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package playbook

import (
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/task/task/checker"
	comm "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/task/task/fs"
	pg "github.com/opencurve/curveadm/internal/task/task/playground"
	"github.com/opencurve/curveadm/internal/tasks"
)

const (
	// checker
	CHECK_TOPOLOGY int = iota
	CHECK_SSH_CONNECT
	CHECK_PERMISSION
	CHECK_KERNEL_VERSION
	CHECK_KERNEL_MODULE
	CHECK_PORT_IN_USE
	CHECK_DESTINATION_REACHABLE
	START_HTTP_SERVER
	CHECK_NETWORK_FIREWALL
	GET_HOST_DATE
	CHECK_HOST_DATE
	CHECK_DISK_SIZE
	CHECK_CHUNKFILE_POOL
	CHECK_S3
	CLEAN_PRECHECK_ENVIRONMENT

	// common
	PULL_IMAGE
	CREATE_CONTAINER
	SYNC_CONFIG
	START_SERVICE
	START_ETCD
	START_MDS
	START_CHUNKSERVER
	START_SNAPSHOTCLONE
	START_METASERVER
	STOP_SERVICE
	RESTART_SERVICE
	CREATE_PHYSICAL_POOL
	CREATE_LOGICAL_POOL
	UPDATE_TOPOLOGY
	INIT_SERVIE_STATUS
	GET_SERVICE_STATUS
	CLEAN_SERVICE
	INIT_SUPPORT
	COLLECT_REPORT
	COLLECT_CURVEADM
	COLLECT_SERVICE
	COLLECT_CLIENT
	BACKUP_ETCD_DATA
	CHECK_MDS_ADDRESS
	GET_CLIENT_STATUS
	INSTALL_CLIENT
	UNINSTALL_CLIENT

	// bs
	BALANCE_LEADER
	START_NEBD_SERVICE
	CREATE_VOLUME
	MAP_IMAGE
	UNMAP_IMAGE

	// bs/target
	START_TARGET_DAEMON
	STOP_TARGET_DAEMON
	ADD_TARGET
	DELETE_TARGET
	LIST_TARGETS

	// fs
	CHECK_CLIENT_S3
	MOUNT_FILESYSTEM
	UMOUNT_FILESYSTEM
	FORMAT_CHUNKFILE_POOL
	GET_FORMAT_STATUS

	// polarfs
	DETECT_OS_RELEASE
	INSTALL_POLARFS
	UNINSTALL_POLARFS

	// playground
	CREATE_PLAYGROUND
	INIT_PLAYGROUND
	START_PLAYGROUND
	REMOVE_PLAYGROUND
	GET_PLAYGROUND_STATUS

	// STOP_FORMAT type stop formatting
	STOP_FORMAT

	// unknown
	UNKNOWN
)

func (p *Playbook) createTasks(step *PlaybookStep) (*tasks.Tasks, error) {
	// (1) default tasks execute options
	config, err := NewSmartConfig(step.Configs)
	if err != nil {
		return nil, err
	}

	// (2) set key-value pair for options
	for k, v := range step.Options {
		p.curveadm.MemStorage().Set(k, v)
	}

	// (3) create task one by one and added into tasks
	var t *task.Task
	once := map[string]bool{}
	curveadm := p.curveadm
	ts := tasks.NewTasks()
	for i := 0; i < config.Len(); i++ {
		// only need to execute task once per host
		switch step.Type {
		case CHECK_SSH_CONNECT,
			GET_HOST_DATE,
			PULL_IMAGE:
			host := config.GetDC(i).GetHost()
			if once[host] {
				continue
			}
			once[host] = true
		}

		switch step.Type {
		// checker
		case CHECK_TOPOLOGY:
			t, err = checker.NewCheckTopologyTask(curveadm, nil)
		case CHECK_SSH_CONNECT:
			t, err = checker.NewCheckSSHConnectTask(curveadm, config.GetDC(i))
		case CHECK_PERMISSION:
			t, err = checker.NewCheckPermissionTask(curveadm, config.GetDC(i))
		case CHECK_KERNEL_VERSION:
			t, err = checker.NewCheckKernelVersionTask(curveadm, config.GetDC(i))
		case CHECK_KERNEL_MODULE:
			t, err = checker.NewCheckKernelModuleTask(curveadm, config.GetCC(i))
		case CHECK_PORT_IN_USE:
			t, err = checker.NewCheckPortInUseTask(curveadm, config.GetDC(i))
		case CHECK_DESTINATION_REACHABLE:
			t, err = checker.NewCheckDestinationReachableTask(curveadm, config.GetDC(i))
		case START_HTTP_SERVER:
			t, err = checker.NewStartHTTPServerTask(curveadm, config.GetDC(i))
		case CHECK_NETWORK_FIREWALL:
			t, err = checker.NewCheckNetworkFirewallTask(curveadm, config.GetDC(i))
		case GET_HOST_DATE:
			t, err = checker.NewGetHostDate(curveadm, config.GetDC(i))
		case CHECK_HOST_DATE:
			t, err = checker.NewCheckDate(curveadm, nil)
		case CHECK_DISK_SIZE:
			t, err = checker.NewCheckDiskSizeTask(curveadm, config.GetDC(i))
		case CHECK_CHUNKFILE_POOL:
			t, err = checker.NewCheckChunkfilePoolTask(curveadm, config.GetDC(i))
		case CHECK_S3:
			t, err = checker.NewCheckS3Task(curveadm, config.GetDC(i))
		case CHECK_MDS_ADDRESS:
			t, err = checker.NewCheckMdsAddressTask(curveadm, config.GetCC(i))
		case CLEAN_PRECHECK_ENVIRONMENT:
			t, err = checker.NewCleanEnvironmentTask(curveadm, config.GetDC(i))
		// common
		case PULL_IMAGE:
			t, err = comm.NewPullImageTask(curveadm, config.GetDC(i))
		case CREATE_CONTAINER:
			t, err = comm.NewCreateContainerTask(curveadm, config.GetDC(i))
		case SYNC_CONFIG:
			t, err = comm.NewSyncConfigTask(curveadm, config.GetDC(i))
		case START_SERVICE,
			START_ETCD,
			START_MDS,
			START_CHUNKSERVER,
			START_SNAPSHOTCLONE,
			START_METASERVER:
			t, err = comm.NewStartServiceTask(curveadm, config.GetDC(i))
		case STOP_SERVICE:
			t, err = comm.NewStopServiceTask(curveadm, config.GetDC(i))
		case RESTART_SERVICE:
			t, err = comm.NewRestartServiceTask(curveadm, config.GetDC(i))
		case CREATE_PHYSICAL_POOL,
			CREATE_LOGICAL_POOL:
			t, err = comm.NewCreateTopologyTask(curveadm, config.GetDC(i))
		case UPDATE_TOPOLOGY:
			t, err = comm.NewUpdateTopologyTask(curveadm, nil)
		case INIT_SERVIE_STATUS:
			t, err = comm.NewInitServiceStatusTask(curveadm, config.GetDC(i))
		case GET_SERVICE_STATUS:
			t, err = comm.NewGetServiceStatusTask(curveadm, config.GetDC(i))
		case CLEAN_SERVICE:
			t, err = comm.NewCleanServiceTask(curveadm, config.GetDC(i))
		case INIT_SUPPORT:
			t, err = comm.NewInitSupportTask(curveadm, config.GetDC(i))
		case COLLECT_REPORT:
			t, err = comm.NewCollectReportTask(curveadm, config.GetDC(i))
		case COLLECT_CURVEADM:
			t, err = comm.NewCollectCurveAdmTask(curveadm, config.GetDC(i))
		case COLLECT_SERVICE:
			t, err = comm.NewCollectServiceTask(curveadm, config.GetDC(i))
		case COLLECT_CLIENT:
			t, err = comm.NewCollectClientTask(curveadm, config.GetAny(i))
		case BACKUP_ETCD_DATA:
			t, err = comm.NewBackupEtcdDataTask(curveadm, config.GetDC(i))
		case GET_CLIENT_STATUS:
			t, err = comm.NewGetClientStatusTask(curveadm, config.GetAny(i))
		case INSTALL_CLIENT:
			t, err = comm.NewInstallClientTask(curveadm, config.GetCC(i))
		case UNINSTALL_CLIENT:
			t, err = comm.NewUninstallClientTask(curveadm, nil)
		// bs
		case FORMAT_CHUNKFILE_POOL:
			t, err = bs.NewFormatChunkfilePoolTask(curveadm, config.GetFC(i))
		case GET_FORMAT_STATUS:
			t, err = bs.NewGetFormatStatusTask(curveadm, config.GetFC(i))
		case STOP_FORMAT:
			t, err = bs.NewStopFormatTask(curveadm, config.GetFC(i))
		case BALANCE_LEADER:
			t, err = bs.NewBalanceTask(curveadm, config.GetDC(i))
		case START_NEBD_SERVICE:
			t, err = bs.NewStartNEBDServiceTask(curveadm, config.GetCC(i))
		case CREATE_VOLUME:
			t, err = bs.NewCreateVolumeTask(curveadm, config.GetCC(i))
		case MAP_IMAGE:
			t, err = bs.NewMapTask(curveadm, config.GetCC(i))
		case UNMAP_IMAGE:
			t, err = bs.NewUnmapTask(curveadm, nil)
		// bs/target
		case START_TARGET_DAEMON:
			t, err = bs.NewStartTargetDaemonTask(curveadm, config.GetCC(i))
		case STOP_TARGET_DAEMON:
			t, err = bs.NewStopTargetDaemonTask(curveadm, nil)
		case ADD_TARGET:
			t, err = bs.NewAddTargetTask(curveadm, config.GetCC(i))
		case DELETE_TARGET:
			t, err = bs.NewDeleteTargetTask(curveadm, nil)
		case LIST_TARGETS:
			t, err = bs.NewListTargetsTask(curveadm, nil)
		// fs
		case CHECK_CLIENT_S3:
			t, err = checker.NewClientS3ConfigureTask(curveadm, config.GetCC(i))
		case MOUNT_FILESYSTEM:
			t, err = fs.NewMountFSTask(curveadm, config.GetCC(i))
		case UMOUNT_FILESYSTEM:
			t, err = fs.NewUmountFSTask(curveadm, config.GetCC(i))
		// polarfs
		case DETECT_OS_RELEASE:
			t, err = bs.NewDetectOSReleaseTask(curveadm, nil)
		case INSTALL_POLARFS:
			t, err = bs.NewInstallPolarFSTask(curveadm, config.GetCC(i))
		case UNINSTALL_POLARFS:
			t, err = bs.NewUninstallPolarFSTask(curveadm, nil)
		// playground
		case CREATE_PLAYGROUND:
			t, err = pg.NewCreatePlaygroundTask(curveadm, config.GetPGC(i))
		case INIT_PLAYGROUND:
			t, err = pg.NewInitPlaygroundTask(curveadm, config.GetPGC(i))
		case START_PLAYGROUND:
			t, err = pg.NewStartPlaygroundTask(curveadm, config.GetPGC(i))
		case REMOVE_PLAYGROUND:
			t, err = pg.NewRemovePlaygroundTask(curveadm, config.GetAny(i))
		case GET_PLAYGROUND_STATUS:
			t, err = pg.NewGetPlaygroundStatusTask(curveadm, config.GetAny(i))

		default:
			return nil, errno.ERR_UNKNOWN_TASK_TYPE.
				F("task type: %d", step.Type)
		}

		if err != nil {
			return nil, err // already is error code
		} else if t == nil {
			continue
		}

		if config.GetType() == TYPE_CONFIG_DEPLOY { // merge task status into one
			t.SetTid(config.GetDC(i).GetId())
			t.SetPtid(config.GetDC(i).GetParentId())
		}
		ts.AddTask(t)
	}

	return ts, nil
}
