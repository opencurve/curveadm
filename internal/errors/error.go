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
 * Created Date: 2021-12-22
 * Author: Jingli Chen (Wine93)
 */

package errors

import (
	"fmt"
)

type ErrorCode struct {
	code    uint
	message string
}

func (e ErrorCode) Error() string {
	return e.message
}

func (e ErrorCode) Format(args ...interface{}) ErrorCode {
	e.message = fmt.Sprintf(e.message, args...)
	return e
}

var (
	ERR_CONFIGURE_NO_SERVICE = ErrorCode{1001, "no service in topology"}
	ERR_SERVICE_NOT_FOUND    = ErrorCode{2001, "service(id=%s) not found"}
	ERROR_REMOTE_UNREACHED   = ErrorCode{3001, "remote unreached"}
)
