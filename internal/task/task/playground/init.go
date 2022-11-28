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
 * Created Date: 2022-11-07
 * Author: Jingli Chen (Wine93)
 */

package playground

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/playground/script"
	"github.com/opencurve/curveadm/pkg/variable"
)

const (
	DEFAULT_CONFIG_DELIMITER = "="
	ETCD_CONFIG_DELIMITER    = ": "
)

func newMutate(cfg interface{}, delimiter string) step.Mutate {
	var serviceCfg map[string]string
	var variables *variable.Variables
	switch cfg.(type) {
	case *topology.DeployConfig:
		dc := cfg.(*topology.DeployConfig)
		serviceCfg = dc.GetServiceConfig()
		variables = dc.GetVariables()
	case *configure.ClientConfig:
		cc := cfg.(*configure.ClientConfig)
		serviceCfg = cc.GetServiceConfig()
		variables = cc.GetVariables()
	}

	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			return
		}

		// replace config
		v, ok := serviceCfg[strings.ToLower(key)]
		if ok {
			value = v
		}

		// replace variable
		value, err = variables.Rendering(value)
		if err != nil {
			return
		}

		out = fmt.Sprintf("%s%s%s", key, delimiter, value)
		return
	}
}

func checkContainerExist(name string, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if len(*out) > 0 {
			return nil
		}
		return errno.ERR_CONTAINER_ALREADT_REMOVED.
			F("name=%s", name)
	}
}

func prepare(dcs []*topology.DeployConfig, poolset, diskType string) (string, error) {
	pool, err := configure.GenerateDefaultClusterPool(dcs, poolset, diskType)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(pool)
	return string(bytes), err
}

func NewInitPlaygroundTask(curveadm *cli.CurveAdm, cfg *configure.PlaygroundConfig) (*task.Task, error) {
	// new task
	kind := cfg.GetKind()
	name := cfg.GetName()
	subname := fmt.Sprintf("kind=%s name=%s", kind, name)
	t := task.NewTask("Init Playground", subname, nil)
	poolset := curveadm.MemStorage().Get(comm.POOLSET).(string)
	poolsetDisktype := curveadm.MemStorage().Get(comm.POOLSET_DISK_TYPE).(string)

	// add step to task
	var containerId string
	layout := topology.GetCurveBSProjectLayout()
	poolJSONPath := path.Join(layout.ToolsConfDir, "topology.json")
	clusterPoolJson, err := prepare(cfg.GetDeployConfigs(), poolset, poolsetDisktype)
	if err != nil {
		return nil, err
	}

	t.AddStep(&step.ListContainers{ // gurantee container exist
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("name=%s", name),
		Out:         &containerId,
		ExecOptions: execOptions(curveadm),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkContainerExist(name, &containerId),
	})
	for _, dc := range cfg.GetDeployConfigs() {
		delimiter := DEFAULT_CONFIG_DELIMITER
		if dc.GetRole() == topology.ROLE_ETCD {
			delimiter = ETCD_CONFIG_DELIMITER
		}
		for _, conf := range dc.GetProjectLayout().ServiceConfFiles {
			t.AddStep(&step.SyncFile{ // sync service config
				ContainerSrcId:    &containerId,
				ContainerSrcPath:  conf.SourcePath,
				ContainerDestId:   &containerId,
				ContainerDestPath: conf.Path,
				KVFieldSplit:      delimiter,
				Mutate:            newMutate(dc, delimiter),
				ExecOptions:       execOptions(curveadm),
			})
		}
		t.AddStep(&step.SyncFile{ // sync tools config
			ContainerSrcId:    &containerId,
			ContainerSrcPath:  layout.ToolsConfSrcPath,
			ContainerDestId:   &containerId,
			ContainerDestPath: layout.ToolsConfSystemPath,
			KVFieldSplit:      DEFAULT_CONFIG_DELIMITER,
			Mutate:            newMutate(dc, DEFAULT_CONFIG_DELIMITER),
			ExecOptions:       execOptions(curveadm),
		})
	}
	t.AddStep(&step.InstallFile{ // install curvebs/curvefs topology
		ContainerId:       &containerId,
		ContainerDestPath: poolJSONPath,
		Content:           &clusterPoolJson,
		ExecOptions:       execOptions(curveadm),
	})
	for _, conf := range []topology.ConfFile{
		{SourcePath: "/curvebs/conf/client.conf", Path: "/curvebs/nebd/conf/client.conf"},
		{SourcePath: "/curvebs/conf/nebd-server.conf", Path: "/etc/nebd/nebd-server.conf"},
		{SourcePath: "/curvebs/conf/nebd-client.conf", Path: "/etc/nebd/nebd-client.conf"},
	} {
		t.AddStep(&step.SyncFile{ // sync service config
			ContainerSrcId:    &containerId,
			ContainerSrcPath:  conf.SourcePath,
			ContainerDestId:   &containerId,
			ContainerDestPath: conf.Path,
			KVFieldSplit:      DEFAULT_CONFIG_DELIMITER,
			Mutate:            newMutate(cfg.GetClientConfig(), DEFAULT_CONFIG_DELIMITER),
			ExecOptions:       execOptions(curveadm),
		})
	}
	t.AddStep(&step.InstallFile{ // install entrypoint
		ContainerId:       &containerId,
		ContainerDestPath: "/entrypoint.sh",
		Content:           &script.ENTRYPOINT,
		ExecOptions:       execOptions(curveadm),
	})

	return t, nil
}
