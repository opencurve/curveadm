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
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 */

package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/cli/command"
	"github.com/opencurve/curveadm/pkg/log"
)

func Execute() {
	curveadm, err := cli.NewCurveAdm()
	if err != nil {
		fmt.Printf("New curveadm failed: %s", err)
		os.Exit(1)
	}

	cmd := command.NewCurveAdmCommand(curveadm)
	res := cmd.Execute()

	// init audit logger
	now := time.Now().Format("2006-01-02")
	logpath := fmt.Sprintf("%s/curveadm-audit-%s.log", curveadm.LogDir(), now)
	err = log.Init("info", logpath)
	if err != nil {
		fmt.Printf("Init audit logger failed: %s", err)
		os.Exit(1)
	}

	if res != nil {
		log.Error(strings.Join(os.Args[1:], " "), log.Field("result", "failed"))
		os.Exit(1)
	} else {
		log.Info(strings.Join(os.Args[1:], " "), log.Field("result", "success"))
	}
}
