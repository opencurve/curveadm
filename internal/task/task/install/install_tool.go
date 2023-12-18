package install

import (
	"fmt"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/pkg/module"
	"path/filepath"
)

func checkPathExist(path string, sshConfig *module.SSHConfig, curveadm *cli.CurveAdm) error {
	sshClient, err := module.NewSSHClient(*sshConfig)
	if err != nil {
		return errno.ERR_SSH_CONNECT_FAILED.E(err)
	}

	module := module.NewModule(sshClient)
	cmd := module.Shell().Stat(path)
	if _, err := cmd.Execute(curveadm.ExecOptions()); err == nil {
		if pass := tui.ConfirmYes(tui.PromptPathExist(path)); !pass {
			return errno.ERR_CANCEL_OPERATION
		}
	}
	return nil
}

func NewInstallToolTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	layout := dc.GetProjectLayout()
	path := curveadm.MemStorage().Get(comm.KEY_INSTALL_PATH).(string)
	confPath := curveadm.MemStorage().Get(comm.KEY_INSTALL_CONF_PATH).(string)
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if err != nil {
		return nil, err
	}

	if err = checkPathExist(path, hc.GetSSHConfig(), curveadm); err != nil {
		return nil, err
	}
	if err = checkPathExist(confPath, hc.GetSSHConfig(), curveadm); err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s", dc.GetHost())
	t := task.NewTask("Install tool v2", subname, hc.GetSSHConfig())

	t.AddStep(&step.CreateDirectory{
		Paths:       []string{filepath.Dir(path)},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.CopyFromContainer{
		ContainerSrcPath: layout.ToolsV2BinaryPath,
		ContainerId:      containerId,
		HostDestPath:     path,
		ExecOptions:      curveadm.ExecOptions(),
	})
	t.AddStep(&step.CreateDirectory{
		Paths:       []string{filepath.Dir(confPath)},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.CopyFromContainer{
		ContainerSrcPath: layout.ToolsV2ConfSystemPath,
		ContainerId:      containerId,
		HostDestPath:     confPath,
		ExecOptions:      curveadm.ExecOptions(),
	})

	return t, nil
}
