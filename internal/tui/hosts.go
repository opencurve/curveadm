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
 * Created Date: 2022-07-21
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package tui

import (
	"strconv"
	"strings"

	configure "github.com/opencurve/curveadm/internal/configure/hosts"
	"github.com/opencurve/curveadm/internal/tui/common"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	FIELD_LIMIT_LENGTH = 30
)

func FormatHosts(hcs []*configure.HostConfig, verbose bool) string {
	lines := [][]interface{}{}
	title := []string{
		"Host",
		"Hostname",
		"User",
		"Port",
		"Private Key File",
		"Forward Agent",
		"Become User",
		"Labels",
		"Envs",
	}
	first, second := tuicommon.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	for i := 0; i < len(hcs); i++ {
		hc := hcs[i]

		host := hc.GetHost()
		hostname := hc.GetHostname()
		user := hc.GetUser()
		port := strconv.Itoa(hc.GetSSHPort())
		forwardAgent := utils.Choose(hc.GetForwardAgent(), "Y", "N")
		becomeUser := utils.Choose(len(hc.GetBecomeUser()) > 0, hc.GetBecomeUser(), "-")
		labels := utils.Choose(len(hc.GetLabels()) > 0, strings.Join(hc.GetLabels(), ","), "-")
		envs := utils.Choose(len(hc.GetEnvs()) > 0, strings.Join(hc.GetEnvs(), ","), "-")
		privateKeyFile := hc.GetPrivateKeyFile()
		if len(privateKeyFile) == 0 {
			privateKeyFile = "-"
		} else if !verbose && len(hc.GetPrivateKeyFile()) > FIELD_LIMIT_LENGTH {
			privateKeyFile = privateKeyFile[:FIELD_LIMIT_LENGTH] + "..."
		}

		lines = append(lines, []interface{}{
			host,
			hostname,
			user,
			port,
			privateKeyFile,
			forwardAgent,
			becomeUser,
			labels,
			envs,
		})
	}

	return common.FixedFormat(lines, 2)
}
