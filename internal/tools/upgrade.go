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
 * Created Date: 2021-11-23
 * Author: Jingli Chen (Wine93)
 */

package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/opencurve/curveadm/cli/cli"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	URL_LATEST_VERSION   = "http://curveadm.nos-eastchina1.126.net/release/__version"
	URL_INSTALL_SCRIPT   = "http://curveadm.nos-eastchina1.126.net/script/install.sh"
	HEADER_VERSION       = "X-Nos-Meta-Curveadm-Version"
	ENV_CURVEADM_UPGRADE = "CURVEADM_UPGRADE"
	ENV_CURVEADM_VERSION = "CURVEADM_VERSION"
)

func calcVersion(v string) int {
	num := 0
	base := 1000
	items := strings.Split(v, ".")
	for _, item := range items {
		n, err := strconv.Atoi(item)
		if err != nil {
			return -1
		}
		num = num*base + n
	}
	return num
}

func isLatest(currentVersion, remoteVersion string) (error, bool) {
	v1 := calcVersion(currentVersion)
	v2 := calcVersion(remoteVersion)
	if v1 == -1 || v2 == -1 {
		return fmt.Errorf("invalid version format: %s, %s", currentVersion, remoteVersion), false
	}

	return nil, v1 >= v2
}

func getLatestVersion(curveadm *cli.CurveAdm) (string, error) {
	version := os.Getenv(ENV_CURVEADM_VERSION)
	if len(version) > 0 {
		return version, nil
	}

	// get latest version from remote
	client := resty.New()
	resp, err := client.R().Head(URL_LATEST_VERSION)
	if err != nil {
		return "", err
	}

	v, ok := resp.Header()[HEADER_VERSION]
	if !ok {
		return "", fmt.Errorf("response header '%s' not exist", HEADER_VERSION)
	} else if err, yes := isLatest(cli.Version, strings.TrimPrefix(v[0], "v")); err != nil {
		return "", err
	} else if yes {
		curveadm.WriteOut("The current version is up-to-date\n")
		return "", nil
	}
	return v[0], nil
}

func Upgrade(curveadm *cli.CurveAdm) error {
	version, err := getLatestVersion(curveadm)
	if err != nil {
		return err
	} else if len(version) == 0 {
		return nil
	} else if pass := tui.ConfirmYes("Upgrade curveadm to %s?", version); !pass {
		return nil
	}

	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("curl -fsSL %s | bash", URL_INSTALL_SCRIPT))
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=true", ENV_CURVEADM_UPGRADE))
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", ENV_CURVEADM_VERSION, version))
	cmd.Stderr = curveadm.Err()
	cmd.Stdout = curveadm.Out()
	return cmd.Run()
}
