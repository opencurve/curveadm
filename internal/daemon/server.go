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

package daemon

import (
	"github.com/opencurve/curveadm/internal/daemon/core"
	"github.com/opencurve/curveadm/internal/daemon/manager"
	"github.com/opencurve/pigeon"
)

func NewServer() *pigeon.HTTPServer {
	server := pigeon.NewHTTPServer("curveadm")
	server.Route("/", manager.Entrypoint)
	server.DefaultRoute(core.Default)
	return server
}
