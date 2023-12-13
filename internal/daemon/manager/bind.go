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
	"mime/multipart"
	"net/http"

	"github.com/opencurve/pigeon"
)

var METHOD_REQUEST map[string]Request

type (
	HandlerFunc func(r *pigeon.Request, ctx *Context) bool

	Context struct {
		Data interface{}
	}

	Request struct {
		httpMethod string
		method     string
		vType      interface{}
		handler    HandlerFunc
	}
)

func init() {
	METHOD_REQUEST = map[string]Request{}
	for _, request := range requests {
		METHOD_REQUEST[request.method] = request
	}
}

type DeployClusterCmdRequest struct {
	Command string `json:"command" binding:"required"`
}

type DeployClusterUploadRequest struct {
	FilePath string                `json:"filepath" form:"filepath" binding:"required"`
	File     *multipart.FileHeader `form:"file" binding:"required"`
}

type DeployClusterDownloadRequest struct {
	FilePath string `json:"filepath" form:"filepath" binding:"required"`
}

var requests = []Request{
	{
		http.MethodPost,
		"cluster.deploy.cmd",
		DeployClusterCmdRequest{},
		DeployClusterCmd,
	},
	{
		http.MethodPost,
		"cluster.deploy.upload",
		DeployClusterUploadRequest{},
		DeployClusterUpload,
	},
	{
		http.MethodGet,
		"cluster.deploy.download",
		DeployClusterDownloadRequest{},
		DeployClusterDownload,
	},
}
