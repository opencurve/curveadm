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
 * Created Date: 2022-05-23
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/tui"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type auditOptions struct{}

func NewAuditCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options auditOptions

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAudit(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runAudit(curveadm *cli.CurveAdm, options auditOptions) error {
	auditLogs, err := curveadm.Storage().GetAuditLogs()
	if err != nil {
		return err
	}

	output := tui.FormatAuditLogs(auditLogs)
	curveadm.WriteOut(output)
	return nil
}
