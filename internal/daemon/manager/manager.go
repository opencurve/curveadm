/*
*  Copyright (c) 2023 NetEase Inc.
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

package manager

import (
	"io"
	"os/exec"

	"github.com/opencurve/curveadm/internal/daemon/core"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/pigeon"
)

func DeployClusterCmd(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*DeployClusterCmdRequest)
	r.Logger().Info("DeployClusterCmd", pigeon.Field("command", data.Command))
	cmd := exec.Command("/bin/bash", "-c", data.Command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.Logger().Warn("DeployClusterCmd failed when execute command",
			pigeon.Field("error", err))
		return core.ExitFailWithData(r, string(out), string(out))
	}
	r.Logger().Info("DeployClusterCmd", pigeon.Field("result", out))
	return core.ExitSuccessWithData(r, string(out))
}

func DeployClusterUpload(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*DeployClusterUploadRequest)
	r.Logger().Info("DeployClusterUpload", pigeon.Field("file", data.FilePath))
	mf, err := data.File.Open()
	if err != nil {
		r.Logger().Warn("DeployClusterUpload failed when open file",
			pigeon.Field("error", err))
		return core.ExitFailWithData(r, err.Error(), err.Error())
	}
	defer mf.Close()
	content, err := io.ReadAll(mf)
	if err != nil {
		r.Logger().Warn("DeployClusterUpload failed when read file",
			pigeon.Field("error", err))
		return core.ExitFailWithData(r, err.Error(), err.Error())
	}
	err = utils.WriteFile(data.FilePath, string(content), 0644)
	if err != nil {
		r.Logger().Warn("DeployClusterUpload failed when write file",
			pigeon.Field("error", err))
		return core.ExitFailWithData(r, err.Error(), err.Error())
	}
	return core.Exit(r, err)
}

func DeployClusterDownload(r *pigeon.Request, ctx *Context) bool {
	data := ctx.Data.(*DeployClusterDownloadRequest)
	r.Logger().Info("DeployClusterDownload", pigeon.Field("file", data.FilePath))
	return r.SendFile(data.FilePath)
}
