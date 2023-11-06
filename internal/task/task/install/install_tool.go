package install

import (
	"fmt"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"strings"
)

const (
	FORMAT_TOOLSV2_CONF_BS = `
global:
  httpTimeout: 500ms
  rpcTimeout: 500ms
  rpcRetryTimes: 1
  maxChannelSize: 4
  showError: false
curvebs:
  mdsAddr: %s
  mdsDummyAddr: %s
  etcdAddr: %s
  snapshotAddr: %s
  snapshotDummyAddr: %s
  root:
    user: root
    password: root_password
`
	FORMAT_TOOLSV2_CONF_FS = `
global:
  httpTimeout: 500ms
  rpcTimeout: 500ms
  rpcRetryTimes: 1
  maxChannelSize: 4
  showError: false
curvefs:
  mdsAddr: %s
  mdsDummyAddr: %s
  etcdAddr: %s
  s3:
    ak: ak
    sk: sk
    endpoint: http://localhost:9000
    bucketname: bucketname
    blocksize: 4 mib
    chunksize: 64 mib
`
)

func genToolV2Conf(dcs []*topology.DeployConfig) string {
	var etcdAddr []string
	var mdsAddr []string
	var mdsDummyAddr []string
	var snapshotCloneDummyAddr []string
	var snapshotCloneProxyAddr []string

	for _, dc := range dcs {
		ip := dc.GetListenIp()
		switch dc.GetRole() {
		case topology.ROLE_ETCD:
			etcdAddr = append(etcdAddr, fmt.Sprintf("%s:%d", ip, dc.GetListenClientPort()))
		case topology.ROLE_MDS:
			mdsAddr = append(mdsAddr, fmt.Sprintf("%s:%d", ip, dc.GetListenPort()))
			mdsDummyAddr = append(mdsDummyAddr, fmt.Sprintf("%s:%d", ip, dc.GetListenDummyPort()))
		case topology.ROLE_SNAPSHOTCLONE:
			snapshotCloneDummyAddr = append(snapshotCloneDummyAddr, fmt.Sprintf("%s:%d", ip, dc.GetListenDummyPort()))
			snapshotCloneProxyAddr = append(snapshotCloneProxyAddr, fmt.Sprintf("%s:%d", ip, dc.GetListenProxyPort()))
		}
	}

	var toolConf string
	mdsAddrStr := strings.Join(mdsAddr, ",")
	mdsDummyAddrStr := strings.Join(mdsDummyAddr, ",")
	etcdAddrStr := strings.Join(etcdAddr, ",")
	snapshotCloneDummyAddrStr := strings.Join(snapshotCloneDummyAddr, ",")
	snapshotCloneProxyAddrStr := strings.Join(snapshotCloneProxyAddr, ",")

	if dcs[0].GetKind() == topology.KIND_CURVEBS {
		toolConf = fmt.Sprintf(strings.TrimSpace(FORMAT_TOOLSV2_CONF_BS), mdsAddrStr, mdsDummyAddrStr, etcdAddrStr, snapshotCloneProxyAddrStr, snapshotCloneDummyAddrStr)
	} else if dcs[0].GetKind() == topology.KIND_CURVEFS {
		toolConf = fmt.Sprintf(strings.TrimSpace(FORMAT_TOOLSV2_CONF_FS), mdsAddrStr, mdsDummyAddrStr, etcdAddrStr)
	}

	return toolConf
}

func NewInstallToolTask(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig) (*task.Task, error) {
	host := curveadm.MemStorage().Get(comm.KEY_CLIENT_HOST).(string)
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s", host)
	t := task.NewTask("Install tool v2", subname, hc.GetSSHConfig())

	confContent := genToolV2Conf(dcs)

	t.AddStep(&step.CreateDirectory{
		Paths:       []string{"~/.curve"},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Curl{
		Url:         "https://curve-tool.nos-eastchina1.126.net/release/curve-latest",
		Insecure:    true,
		Silent:      true,
		Output:      "/tmp/curve-latest",
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Chmod{
		Mode:        "+x",
		File:        "/tmp/curve-latest",
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.MoveFile{
		Source:      "/tmp/curve-latest",
		Dest:        "/usr/bin/curve",
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{
		Content:      &confContent,
		HostDestPath: "~/.curve/curve.yaml",
		ExecOptions:  curveadm.ExecOptions(),
	})

	return t, nil
}
