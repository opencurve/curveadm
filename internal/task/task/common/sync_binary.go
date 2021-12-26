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
 * Created Date: 2021-11-25
 * Author: Hailang Mo (wavemomo)
 */

package common

/*
import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/opencurve/curveadm/internal/task"
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/context"
)

const KEY_BINARY_PATH = "BINARY_PATH"

type step2SyncBinary struct {
	containerID           string
	remoteContainerBinary string
	localBinary           string
}

func (s *step2SyncBinary) Execute(ctx *context.Context) error {
	remoteHostBinaryPath := fmt.Sprintf("/tmp/%s", path.Base(s.localBinary))
	if err := ctx.Module().Scp(s.localBinary, remoteHostBinaryPath); err != nil {
		return err
	}

	_, err := ctx.Module().SshShell("sudo docker cp %s %s:%s", remoteHostBinaryPath, s.containerID,
		s.remoteContainerBinary)
	return err

}

func (s *step2SyncBinary) Rollback(ctx *context.Context) {}

func NewSyncBinaryTask(curveadm *cli.CurveAdm, dc *configure.DeployConfig) (*task.Task, error) {
	serviceId := configure.GetServiceId(curveadm.ClusterId(), dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	binaryPath := curveadm.MemStorage().Get(KEY_BINARY_PATH).(string)
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return nil, err
	}
	subname := fmt.Sprintf("host=%s role=%s binary=%s", dc.GetHost(), dc.GetRole(), absPath)
	t := task.NewTask("Sync Binary", subname, dc)
	remotePath := fmt.Sprintf("/usr/local/curvefs/%s/sbin/%s", dc.GetRole(), path.Base(binaryPath))
	t.AddStep(&step2SyncBinary{
		containerID:           containerId,
		localBinary:           binaryPath,
		remoteContainerBinary: remotePath,
	})
	return t, nil
}
*/
