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
 * Author: Jingli Chen (Wine93)
 */

package scripts

const (
	STATUS_OK      = "CURVEADM_OK"
	STATUS_FAIL    = "CURVEADM_FAIL"
	STATUS_TIMEOUT = "CURVEADM_TIMEOUT"
)

var (
	SCRIPT_WAIT              string = WAIT
	SCRIPT_COLLECT           string = COLLECT
	SCRIPT_REPORT            string = REPORT
	SCRIPT_FORMAT            string = FORMAT
	SCRIPT_MAP               string = MAP
	SCRIPT_TARGET            string = TARGET
	SCRIPT_RECYCLE           string = RECYCLE
	SCRIPT_CREATEFS          string = CREATEFS
	SCRIPT_CREATE_VOLUME     string = CREATE_VOLUME
	SCRIPT_WAIT_CHUNKSERVERS string = WAIT_CHUNKSERVERS
	SCRIPT_START_NGINX       string = START_NGINX
)
