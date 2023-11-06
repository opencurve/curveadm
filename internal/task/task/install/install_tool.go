package install

import (
	"fmt"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

func NewInstallToolTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	layout := dc.GetProjectLayout()
	host := curveadm.MemStorage().Get(comm.KEY_CLIENT_HOST).(string)
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s", host)
	t := task.NewTask("Install tool v2", subname, hc.GetSSHConfig())

	t.AddStep(&step.CopyFromContainer{
		ContainerSrcPath: layout.ToolsV2BinaryPath,
		ContainerId:      containerId,
		HostDestPath:     "/usr/local/bin/curve",
		ExecOptions:      curveadm.ExecOptions(),
	})
	t.AddStep(&step.CreateDirectory{
		Paths:       []string{"~/.curve"},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.CopyFromContainer{
		ContainerSrcPath: layout.ToolsV2ConfSystemPath,
		ContainerId:      containerId,
		HostDestPath:     "~/.curve/curve.yaml",
		ExecOptions:      curveadm.ExecOptions(),
	})

	return t, nil
}
