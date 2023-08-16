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
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

type (
	step2EncryptFile struct {
		source string
		dest   string
		secret string
	}
)

func (s *step2EncryptFile) Execute(ctx *context.Context) error {
	err := utils.EncryptFile(s.source, s.dest, s.secret)
	if err != nil {
		return errno.ERR_ENCRYPT_FILE_FAILED.E(err)
	}
	return nil
}

func NewCollectReportTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	/*
		// new task
		kind := dc.GetKind()
		subname := fmt.Sprintf("cluster=%s kind=%s",
			curveadm.ClusterName(), kind)
		t := task.NewTask("Collect Report", subname, nil)

		// add step to task
		var out, content string
		var outs []string
		var success bool
		secret := curveadm.MemStorage().Get(comm.KEY_SECRET).(string)
		urlFormat := curveadm.MemStorage().Get(comm.KEY_SUPPORT_UPLOAD_URL_FORMAT).(string)
		baseDir := TEMP_DIR
		vname := utils.NewVariantName(fmt.Sprintf("report_%s", utils.RandString(5)))
		localPath := path.Join(baseDir, vname.Name)                // /tmp/report_is90x
		localTarballPath := path.Join(baseDir, vname.CompressName) // /tmp/report_is90x.tar.gz
		localEncryptdTarballPath := path.Join(baseDir, vname.EncryptCompressName)
		httpSavePath := path.Join("/", encodeSecret(secret), "report")
		options := curveadm.ExecOptions()
		options.ExecWithSudo = false
		options.ExecInLocal = true
		commands := []string{
			fmt.Sprintf("%s hosts ls -v", curveadm.BinPath()),
			fmt.Sprintf("%s config show", curveadm.BinPath()),
			fmt.Sprintf("%s status -sv", curveadm.BinPath()),
			fmt.Sprintf("%s client status", curveadm.BinPath()),
		}

		for _, command := range commands {
			t.AddStep(&step.Command{
				Command:     command,
				Success:     &success,
				Out:         &out,
				ExecOptions: options,
			})
			t.AddStep(&step.Lambda{
				Lambda: appendOut(command, &success, &out, &outs),
			})
		}
		t.AddStep(&step.Lambda{
			Lambda: convert2Content(&outs, &content),
		})
		t.AddStep(&step.InstallFile{
			Content:      &content,
			HostDestPath: localPath,
			ExecOptions:  options,
		})
		t.AddStep(&step.Tar{
			File:        localPath,
			Archive:     localTarballPath,
			Create:      true,
			Gzip:        true,
			Verbose:     true,
			ExecOptions: options,
		})
		t.AddStep(&step2EncryptFile{
			source: localTarballPath,
			dest:   localEncryptdTarballPath,
			secret: secret,
		})
		t.AddStep(&step.Curl{ // upload to curve team // curl -F "path=@$FILE" http://localhost:8080/upload\?path\=/
			Url:         fmt.Sprintf(urlFormat, httpSavePath),
			Form:        fmt.Sprintf("path=@%s", localEncryptdTarballPath),
			ExecOptions: options,
		})
		t.AddPostStep(&step.RemoveFile{
			Files:       []string{localPath, localTarballPath, localEncryptdTarballPath},
			ExecOptions: options,
		})

		return t, nil
	*/
	return nil, nil
}
