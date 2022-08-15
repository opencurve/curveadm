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

package tasks

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/plugin"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"

)

/*
 * playbook
 * ├── tasks1 (e.g.: pull image)
 * ├── tasks2 (e.g.: create container)
 * ├── ...
 * └── tasksn (e.g.: start container)
 *     ├── task1 (e.g.: start container in host1)
 *     ├── task2 (e.g.: start container in host2)
 *     ├── ...
 *     └── taskn (e.g.: start container in host3)
 *         ├── step1 (e.g: start container)
 *         ├── step2 (e.g: check container status)
 *         ├── ...
 *         └── stepn (e.g: start crotab iff status is ok)
 *
 * tasks are made up of many same type tasks which only executed in different hosts or roles
 */
const (
	TYPE_CONFIG_DEPLOY int = iota
	TYPE_CONFIG_CLIENT
	TYPE_CONFIG_FORMAT
	TYPE_CONFIG_PLUGIN
	TYPE_CONFIG_PLAYGROUND
	TYPE_CONFIG_NULL
)

type configs struct {
	ctype  int
	length int
	dcs    []*topology.DeployConfig
	ccs    []*configure.ClientConfig
	fcs    []*configure.FormatConfig
	pcs    []*plugin.PluginConfig
	pgcs   []*configure.PlaygroundConfig
}

const (
	// checker
	CHECK_TOPOLOGY int = iota
	CHECK_SSH_CONNECT
	CHECK_PERMISSION
	CHECK_KERNEL_VERSION
	CHECK_PORT_IN_USE
	CHECK_DESTINATION_REACHABLE
	CHECK_NETWORK_FIREWALL
	GET_HOST_DATE
	CHECK_DATE
	CHECK_CHUNKFILE_POOL
	CHECK_S3

	// common
	PULL_IMAGE
	CREATE_CONTAINER
	SYNC_CONFIG
	START_SERVICE
	STOP_SERVICE
	RESTART_SERVICE
	CREATE_POOL
	GET_SERVICE_STATUS
	CLEAN_SERVICE
	COLLECT_SERVICE
	BACKUP_ETCD_DATA

	// bs
	BALANCE_LEADER
	START_NEBD_SERVICE
	MAP_IMAGE
	UNMAP_IMAGE

	// bs/target
	START_TARGET_DAEMON
	ADD_TARGET
	DELETE_TARGET
	LIST_TARGETS

	// fs
	MOUNT_FILESYSTEM
	UMOUNT_FILESYSTEM
	CHECK_MOUNT_STATUS
	FORMAT_CHUNKFILE_POOL
	GET_FORMAT_STATUS

	// plugin
	RUN_PLUGIN

	// playground
	RUN_PLAYGROUND
	REMOVE_PLAYGROUND

	// unknown
	UNKNOWN
)

func newConfigs(configSlice interface{}) (*configs, error) {
	configs := &configs{
		dcs: []*topology.DeployConfig{},
		ccs: []*configure.ClientConfig{},
		fcs: []*configure.FormatConfig{},
		pcs: []*plugin.PluginConfig{},
	}
	switch configSlice.(type) {
	case []*topology.DeployConfig:
		configs.ctype = TYPE_CONFIG_DEPLOY
		configs.dcs = configSlice.([]*topology.DeployConfig)
		configs.length = len(configs.dcs)
	case []*configure.ClientConfig:
		configs.ctype = TYPE_CONFIG_CLIENT
		configs.ccs = configSlice.([]*configure.ClientConfig)
		configs.length = len(configs.ccs)
	case []*configure.FormatConfig:
		configs.ctype = TYPE_CONFIG_FORMAT
		configs.fcs = configSlice.([]*configure.FormatConfig)
		configs.length = len(configs.fcs)
	case []*plugin.PluginConfig:
		configs.ctype = TYPE_CONFIG_PLUGIN
		configs.pcs = configSlice.([]*plugin.PluginConfig)
		configs.length = len(configs.pcs)
	case *topology.DeployConfig:
		configs.ctype = TYPE_CONFIG_DEPLOY
		configs.dcs = append(configs.dcs, configSlice.(*topology.DeployConfig))
		configs.length = 1
	case *configure.ClientConfig:
		configs.ctype = TYPE_CONFIG_CLIENT
		configs.ccs = append(configs.ccs, configSlice.(*configure.ClientConfig))
		configs.length = 1
	case *configure.FormatConfig:
		configs.ctype = TYPE_CONFIG_FORMAT
		configs.fcs = append(configs.fcs, configSlice.(*configure.FormatConfig))
		configs.length = 1
	case *configure.PlaygroundConfig:
		configs.ctype = TYPE_CONFIG_PLAYGROUND
		configs.pgcs = append(configs.pgcs, configSlice.(*configure.PlaygroundConfig))
		configs.length = 1
	case nil:
		configs.ctype = TYPE_CONFIG_NULL
		configs.length = 1
	default:
		return nil, errno.ERR_UNSUPPORT_CONFIG_TYPE
	}
	return configs, nil
}

func ExecTasks(taskType int, curveadm *cli.CurveAdm, configSlice interface{}) error {
	return nil
	/*
		var t *task.Task
		var dc *topology.DeployConfig
		var cc *configure.ClientConfig
		var fc *configure.FormatConfig
		var pc *plugin.PluginConfig
		var pgc *configure.PlaygroundConfig

		configs, err := newConfigs(configSlice)
		if err != nil {
			return err
		}

		tasks := NewTasks()
		option := ExecOption{
			Concurrency:   10,
			SilentSubBar:  false,
			SilentMainBar: false,
			SkipError:     false,
		}

		// add task into tasks
		pullImage := map[string]bool{}
		breakLoop := false
		ctype := configs.ctype
		for i := 0; i < configs.length; i++ {
			// config type
			switch ctype {
			case TYPE_CONFIG_DEPLOY:
				dc = configs.dcs[i]
			case TYPE_CONFIG_CLIENT:
				cc = configs.ccs[i]
			case TYPE_CONFIG_FORMAT:
				fc = configs.fcs[i]
			case TYPE_CONFIG_PLUGIN:
				pc = configs.pcs[i]
			case TYPE_CONFIG_PLAYGROUND:
				pgc = configs.pgcs[i]
			case TYPE_CONFIG_NULL: // do nothing
			}

			// task type
			switch taskType {
			// checker
			case CHECK_TOPOLOGY:
				t, err = checker.NewCheckTopologyTask(curveadm, configs.dcs)
				breakLoop = true
			case CHECK_SSH_CONNECT:
				t, err = checker.NewCheckSSHConnectTask(curveadm, dc)
			case CHECK_PERMISSION:
				t, err = checker.NewCheckPermissionTask(curveadm, dc)
			case CHECK_KERNEL_VERSION:
				t, err = checker.NewCheckKernelVersionTask(curveadm, dc)
			case CHECK_PORT_IN_USE:
				t, err = checker.NewCheckPortInUseTask(curveadm, dc)
			case CHECK_DESTINATION_REACHABLE:
				t, err = checker.NewCheckDestinationReachableTask(curveadm, dc)
			case CHECK_NETWORK_FIREWALL:
				t, err = checker.NewCheckNetworkFirewall(curveadm, dc)
			case GET_HOST_DATE:
				t, err = checker.NewGetHostDate(curveadm, dc)
			case CHECK_DATE:
				t, err = checker.NewCheckDate(curveadm, dc)
			case CHECK_CHUNKFILE_POOL:
				t, err = checker.NewCheckChunkfilePoolTask(curveadm, dc)
			case CHECK_S3:
				t, err = checker.NewCheckS3Task(curveadm, dc)

			// common
			case PULL_IMAGE:
				// prevent reacheing docker hub pull rate limit
				if pullImage[dc.GetHost()] == true {
					continue
				}
				t, err = comm.NewPullImageTask(curveadm, dc)
				pullImage[dc.GetHost()] = true
			case CREATE_CONTAINER:
				t, err = comm.NewCreateContainerTask(curveadm, dc)
			case SYNC_CONFIG:
				t, err = comm.NewSyncConfigTask(curveadm, dc)
			case START_SERVICE:
				t, err = comm.NewStartServiceTask(curveadm, dc)
			case STOP_SERVICE:
				t, err = comm.NewStopServiceTask(curveadm, dc)
			case RESTART_SERVICE:
				t, err = comm.NewRestartServiceTask(curveadm, dc)
			case CREATE_POOL:
				t, err = comm.NewCreateTopologyTask(curveadm, dc)
			case GET_SERVICE_STATUS:
				option.SilentSubBar = true
				option.SkipError = true
				t, err = comm.NewGetServiceStatusTask(curveadm, dc)
			case CLEAN_SERVICE:
				t, err = comm.NewCleanServiceTask(curveadm, dc)
			case COLLECT_SERVICE:
				t, err = comm.NewCollectServiceTask(curveadm, dc)
			case BACKUP_ETCD_DATA:
				t, err = comm.NewBackupEtcdDataTask(curveadm, dc)

			// bs
			case FORMAT_CHUNKFILE_POOL:
				t, err = bs.NewFormatChunkfilePoolTask(curveadm, fc)
			case GET_FORMAT_STATUS:
				option.SilentSubBar = true
				option.SkipError = true
				t, err = bs.NewGetFormatStatusTask(curveadm, fc)
			case BALANCE_LEADER:
				t, err = bs.NewBalanceTask(curveadm, dc)
			case START_NEBD_SERVICE:
				option.SilentSubBar = true
				t, err = bs.NewStartNEBDServiceTask(curveadm, cc)
			case MAP_IMAGE:
				option.SilentSubBar = true
				t, err = bs.NewMapTask(curveadm, cc)
			case UNMAP_IMAGE:
				option.SilentSubBar = true
				t, err = bs.NewUnmapTask(curveadm, cc)

			// bs/target
			case START_TARGET_DAEMON:
				option.SilentSubBar = true
				t, err = bs.NewStartTargetDaemonTask(curveadm, cc)
			case ADD_TARGET:
				option.SilentSubBar = true
				t, err = bs.NewAddTargetTask(curveadm, cc)
			case DELETE_TARGET:
				option.SilentSubBar = true
				t, err = bs.NewDeleteTargetTask(curveadm, cc)
			case LIST_TARGETS:
				option.SilentSubBar = true
				t, err = bs.NewListTargetsTask(curveadm, cc)

			// fs
			case MOUNT_FILESYSTEM:
				option.SilentSubBar = true
				t, err = fs.NewMountFSTask(curveadm, cc)
			case UMOUNT_FILESYSTEM:
				option.SilentSubBar = true
				t, err = fs.NewUmountFSTask(curveadm, cc)
			case CHECK_MOUNT_STATUS:
				option.SilentMainBar = true
				option.SilentSubBar = true
				t, err = fs.NewGetMountStatusTask(curveadm, cc)

			// plugin
			case RUN_PLUGIN:
				t, err = plg.NewRunPluginTask(curveadm, pc)

			// playground
			case RUN_PLAYGROUND:
				option.SilentSubBar = true
				t, err = pg.NewRunPlaygroundTask(curveadm, pgc)
			case REMOVE_PLAYGROUND:
				option.SilentSubBar = true
				t, err = pg.NewRemovePlaygroundTask(curveadm, pgc)

			default:
				return errno.ERR_UNKNOWN_TASK_TYPE.F("task type: %s", taskType)
			}

			if err != nil {
				return err // error code
			}

			if ctype == TYPE_CONFIG_DEPLOY { // merge task status into one
				t.SetTid(dc.GetId())
				t.SetPtid(dc.GetParentId())
			}
			tasks.AddTask(t)

			if breakLoop {
				break
			}
		}

		return tasks.Execute(option)
	*/
}
