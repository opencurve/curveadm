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
 * Created Date: 2022-08-01
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package hosts

import (
	"fmt"

	comm "github.com/opencurve/curveadm/internal/configure/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	DEFAULT_SSH_PORT  = 22
	DEFAULT_HTTP_PORT = 8000
	SSH_PROTOCOL      = "ssh"
	HTTP_PROTOCOL     = "http"
)

var (
	itemset = comm.NewItemSet()

	CONFIG_HOST = itemset.Insert(
		"host",
		comm.REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_HOSTNAME = itemset.Insert(
		"hostname",
		comm.REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_SSH_HOSTNAME = itemset.Insert(
		"ssh_hostname",
		comm.REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_USER = itemset.Insert(
		"user",
		comm.REQUIRE_STRING,
		false,
		func(hc *HostConfig) interface{} {
			return utils.GetCurrentUser()
		},
	)

	CONFIG_SSH_PORT = itemset.Insert(
		"ssh_port",
		comm.REQUIRE_POSITIVE_INTEGER,
		false,
		DEFAULT_SSH_PORT,
	)

	CONFIG_PRIVATE_CONFIG_FILE = itemset.Insert(
		"private_key_file",
		comm.REQUIRE_STRING,
		false,
		func(hc *HostConfig) interface{} {
			return fmt.Sprintf("%s/.ssh/id_rsa", utils.GetCurrentHomeDir())
		},
	)

	CONFIG_FORWARD_AGENT = itemset.Insert(
		"forward_agent",
		comm.REQUIRE_BOOL,
		false,
		false,
	)

	CONFIG_BECOME_USER = itemset.Insert(
		"become_user",
		comm.REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_PROTOCOL = itemset.Insert(
		"protocol",
		comm.REQUIRE_STRING,
		false,
		SSH_PROTOCOL,
	)

	CONFIG_HTTP_PORT = itemset.Insert(
		"http_port",
		comm.REQUIRE_POSITIVE_INTEGER,
		false,
		DEFAULT_HTTP_PORT,
	)
)
