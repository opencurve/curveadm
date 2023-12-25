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

import (
	_ "embed"
)

const (
	STATUS_OK      = "CURVEADM_OK"
	STATUS_FAIL    = "CURVEADM_FAIL"
	STATUS_TIMEOUT = "CURVEADM_TIMEOUT"
)

var (
	// Common

	//go:embed shell/wait.sh
	WAIT string
	//go:embed shell/report.sh
	REPORT string
	//go:embed shell/add_etcd.sh
	ADD_ETCD string
	//go:embed shell/remove_etcd.sh
	REMOVE_ETCD string

	// CurveBS

	//go:embed shell/format.sh
	FORMAT string
	//go:embed shell/wait_chunkserver.sh
	WAIT_CHUNKSERVERS string
	//go:embed shell/start_nginx.sh
	START_NGINX string
	//go:embed shell/create_volume.sh
	CREATE_VOLUME string
	//go:embed shell/map.sh
	MAP string
	//go:embed shell/target.sh
	TARGET string
	//go:embed shell/recycle.sh
	RECYCLE string

	// CurveFS

	//go:embed shell/create_fs.sh
	CREATE_FS string

	//go:embed shell/mark_server_pendding.sh
	MARK_SERVER_PENDDING string
)
