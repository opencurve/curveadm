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
 * Created Date: 2021-11-26
 * Author: Jingli Chen (Wine93)
 */

package common

import (
	"fmt"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"path"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

func NewCollectClientTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	client := v.(storage.Client)
	hc, err := curveadm.GetHost(client.Host)
	if err != nil {
		return nil, err
	}

	// new task
	containerId := client.ContainerId
	subname := fmt.Sprintf("host=%s kind=%s containerId=%s",
		hc.GetHost(), client.Kind, tui.TrimContainerId(containerId))
	t := task.NewTask("Collect Client", subname, hc.GetSSHConfig(), hc.GetHttpConfig())

	// add step to task
	var out string
	clientId := client.Id
	secret := curveadm.MemStorage().Get(comm.KEY_SECRET).(string)
	urlFormat := curveadm.MemStorage().Get(comm.KEY_SUPPORT_UPLOAD_URL_FORMAT).(string)
	baseDir := TEMP_DIR
	vname := utils.NewVariantName(fmt.Sprintf("%s_%s", clientId, utils.RandString(5)))
	remoteSaveDir := fmt.Sprintf("/%s/%s", baseDir, vname.Name)               // /tmp/7b510fb63730_ox1fe
	remoteTarbllPath := path.Join(baseDir, vname.CompressName)                // /tmp/7b510fb63730_ox1fe.tar.gz
	localTarballPath := path.Join(baseDir, vname.CompressName)                // /tmp/7b510fb63730_ox1fe.tar.gz
	localEncryptdTarballPath := path.Join(baseDir, vname.EncryptCompressName) // // /tmp/7b510fb63730_ox1fe-encrypted.tar.gz
	httpSavePath := path.Join("/", encodeSecret(secret), "client")            // /34701feb224479a20a5090510f648037/client
	containerLogDir := utils.Choose(client.Kind == topology.KIND_CURVEBS,
		"/curvebs/nebd/logs", "/curvefs/client/logs")
	containerConfDir := utils.Choose(client.Kind == topology.KIND_CURVEBS,
		"/curvebs/nebd/conf", "/curvefs/client/conf")
	localOptions := curveadm.ExecOptions()
	localOptions.ExecInLocal = true
	t.AddStep(&step.CreateDirectory{
		Paths:       []string{remoteSaveDir},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2CopyFilesFromContainer{ // copy logs directory
		containerId: containerId,
		files:       &[]string{containerLogDir},
		hostDestDir: remoteSaveDir,
		curveadm:    curveadm,
	})
	t.AddStep(&step2CopyFilesFromContainer{ // copy conf directory
		containerId: containerId,
		files:       &[]string{containerConfDir},
		hostDestDir: remoteSaveDir,
		curveadm:    curveadm,
	})
	t.AddStep(&step.ContainerLogs{
		ContainerId: containerId,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{
		Content:      &out,
		HostDestPath: fmt.Sprintf("%s/docker.log", path.Join(remoteSaveDir, "logs")),
		ExecOptions:  curveadm.ExecOptions(),
	})
	t.AddStep(&step.Tar{
		File:        remoteSaveDir,
		Archive:     remoteTarbllPath,
		Create:      true,
		Gzip:        true,
		Verbose:     true,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.DownloadFile{
		RemotePath:  remoteTarbllPath,
		LocalPath:   localTarballPath,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2EncryptFile{
		source: localTarballPath,
		dest:   localEncryptdTarballPath,
		secret: secret,
	})
	t.AddStep(&step.Curl{ // upload to curve team // curl -F "path=@$FILE" http://localhost:8080/upload\?path\=/
		Url:         fmt.Sprintf(urlFormat, httpSavePath),
		Form:        fmt.Sprintf("path=@%s", localEncryptdTarballPath),
		ExecOptions: localOptions,
	})
	t.AddPostStep(&step.RemoveFile{
		Files:       []string{remoteSaveDir, remoteTarbllPath},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddPostStep(&step.RemoveFile{
		Files:       []string{localTarballPath, localEncryptdTarballPath},
		ExecOptions: localOptions,
	})

	return t, nil
}
