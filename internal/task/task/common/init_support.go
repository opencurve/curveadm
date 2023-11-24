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
 * Created Date: 2022-08-14
 * Author: Jingli Chen (Wine93)
 */

package common

import (
	"fmt"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"path"
)

func NewInitSupportTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	// new task
	kind := dc.GetKind()
	subname := fmt.Sprintf("cluster=%s kind=%s",
		curveadm.ClusterName(), kind)
	t := task.NewTask("Init Support", subname, nil, nil)

	/*
	 * 0d7a7103521da69c6331a96355142c3b
	 *   data
	 *     curveadm_db-encrypted.tar.gz
	 *   report
	 *     report-encrypted.tar.gz (curveadm hosts ls)
	 *   service
	 *     etcd
	 *       7b510fb63730-encrypted.tar.gz
	 *       978333085318-encrypted.tar.gz
	 *     mds
	 *       ...
	 *   client
	 *     362d538778ad-encrypted.tar.gz
	 *     b0d56cfaad14-encrypted.tar.gz
	 */
	roles := topology.CURVEBS_ROLES
	if kind == topology.KIND_CURVEFS {
		roles = topology.CURVEFS_ROLES
	}
	secret := curveadm.MemStorage().Get(comm.KEY_SECRET).(string)
	urlFormat := curveadm.MemStorage().Get(comm.KEY_SUPPORT_UPLOAD_URL_FORMAT).(string)

	options := curveadm.ExecOptions()
	options.ExecInLocal = true
	root := encodeSecret(secret)
	dirs := []string{
		root,
		path.Join(root, "data"),
		path.Join(root, "report"),
		path.Join(root, "service"),
		path.Join(root, "client"),
	}
	for _, role := range roles {
		dirs = append(dirs, path.Join(root, "service", role))
	}
	// curl -F "mkdir=$DIR_NAME" http://localhost:8080/upload\?path=\/
	for _, dir := range dirs {
		t.AddStep(&step.Curl{
			Url:         fmt.Sprintf(urlFormat, "/"),
			Form:        fmt.Sprintf("mkdir=%s", dir),
			Insecure:    true,
			Silent:      true,
			ExecOptions: options,
		})
	}

	return t, nil
}
