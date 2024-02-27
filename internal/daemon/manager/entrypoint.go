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

/*
* Project: CurveAdm
* Created Date: 2023-12-13
* Author: liuminjian
 */

package manager

import (
	"reflect"

	"github.com/mcuadros/go-defaults"
	"github.com/opencurve/curveadm/internal/daemon/core"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/pigeon"
)

func Entrypoint(r *pigeon.Request) bool {
	if r.Method != pigeon.HTTP_METHOD_GET &&
		r.Method != pigeon.HTTP_METHOD_POST {
		return core.Exit(r, errno.ERR_UNSUPPORT_HTTP_METHOD)
	}

	request, ok := METHOD_REQUEST[r.Args["method"]]
	if !ok {
		return core.Exit(r, errno.ERR_UNSUPPORT_METHOD_ARGUMENT)
	} else if request.httpMethod != r.Method {
		return core.Exit(r, errno.ERR_HTTP_METHOD_MISMATCHED)
	}

	vType := reflect.TypeOf(request.vType)
	data := reflect.New(vType).Interface()
	if err := r.BindBody(data); err != nil {
		r.Logger().Error("bad request form param",
			pigeon.Field("error", err))
		return core.Exit(r, errno.ERR_BAD_REQUEST_FORM_PARAM)
	}
	defaults.SetDefaults(data)
	return request.handler(r, &Context{data})
}
