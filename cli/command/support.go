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
 * Created Date: 2021-11-26
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tasks"
	"github.com/opencurve/curveadm/internal/tools"
	"github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	PROMPT = `FYI:
  > We have collected logs for troubleshooting,
  > and now we will send these logs to the curve center.
  > Please don't worry about the data security,
  > we guarantee that all logs are encrypted 
  > and only you have the secret key.
`
)

type supportOptions struct {
}

func NewSupportCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options supportOptions

	cmd := &cobra.Command{
		Use:   "support",
		Short: "Get support from Curve team",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSupport(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func collectService(curveadm *cli.CurveAdm, saveDir string) error {
	dcs, err := configure.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	}

	memStorage := curveadm.MemStorage()
	memStorage.Set(tasks.KEY_COLLECT_SAVE_DIR, saveDir)
	if err := tasks.ExecParallelTasks(tasks.COLLECT_SERVICE, curveadm, dcs); err != nil {
		return curveadm.NewPromptError(err, "")
	}
	return nil
}

func collectCurveadm(curveadm *cli.CurveAdm, saveDir string) error {
	saveCurveAdmDir := fmt.Sprintf("%s/curveadm", saveDir)
	if _, err := utils.ExecShell("mkdir -p %s", saveCurveAdmDir); err != nil {
		return err
	}

	for _, item := range []string{curveadm.DataDir(), curveadm.LogDir()} {
		if _, err := utils.ExecShell("cp -r %s %s", item, saveCurveAdmDir); err != nil {
			return err
		}
	}
	return nil
}

func upload2CurveTeam(saveDir, secretKey string) error {
	rootDir := filepath.Dir(saveDir)                   // ~/.curveadm/temp/curveadm-support-8589bddc3c2c56eedfdb0fc2194b3ecd
	filename := filepath.Base(saveDir)                 // curveadm-support-8589bddc3c2c56eedfdb0fc2194b3ecd
	tarball := filename + ".tar.gz"                    // curveadm-support-8589bddc3c2c56eedfdb0fc2194b3ecd.tar.gz
	encryptedTarball := filename + "-encrypted.tar.gz" //curveadm-support-8589bddc3c2c56eedfdb0fc2194b3ecd-encrypted.tar.gz
	srcfile := fmt.Sprintf("%s/%s", rootDir, tarball)
	dstfile := fmt.Sprintf("%s/%s", rootDir, encryptedTarball)
	cmd := fmt.Sprintf("cd %s; tar -zcvf %s %s", rootDir, tarball, filename)
	if _, err := utils.ExecShell(cmd); err != nil {
		return err
	} else if err := utils.EncryptFile(srcfile, dstfile, secretKey); err != nil {
		return err
	} else if err := tools.Upload(dstfile); err != nil {
		return err
	}

	return nil
}

/*
 * curveadm-support-8589bddc3c2c56eedfdb0fc2194b3ecd
 *   1_etcd_10.0.0.1_1
 *     logs/
 *     bin/
 *     conf/
 *     core/
 *   curveadm
 *     logs/
 *     data/
 */
func runSupport(curveadm *cli.CurveAdm, options supportOptions) error {
	secretKey := utils.RandString(32)
	saveDir := fmt.Sprintf("%s/curveadm-support-%s", curveadm.TempDir(), utils.MD5Sum(secretKey))
	if err := os.Mkdir(saveDir, 0755); err != nil {
		return err
	}
	defer func() {
		utils.ExecShell("rm -rf %s*", saveDir)
	}()

	if err := collectService(curveadm, saveDir); err != nil {
		return err
	} else if err := collectCurveadm(curveadm, saveDir); err != nil {
		return err
	}

	curveadm.WriteOut(color.YellowString(PROMPT))
	if pass := common.ConfirmYes("Do you want to continue? [y/N]: "); !pass {
		return nil
	}

	if err := upload2CurveTeam(saveDir, secretKey); err != nil {
		return err
	}
	curveadm.WriteOut("Secret Key: %s\n", color.GreenString(secretKey))
	return nil
}
