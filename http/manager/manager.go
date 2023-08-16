/*
*  Copyright (c) 2023 NetEase Inc.
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
* Project: Curveadm
* Created Date: 2023-04-04
* Author: wanghai (SeanHai)
 */

package manager

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/cli/command"
	"github.com/opencurve/curveadm/cli/command/cluster"
	"github.com/opencurve/curveadm/cli/command/config"
	"github.com/opencurve/curveadm/cli/command/disks"
	"github.com/opencurve/curveadm/cli/command/hosts"
	"github.com/opencurve/curveadm/cli/command/monitor"
	"github.com/opencurve/curveadm/http/core"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/pigeon"
)

const (
	KEY_ETCD_ENDPOINT                 = "etcd.address"
	KEY_MDS_ENDPOINT                  = "mds.address"
	KEY_MDS_DUMMY_ENDPOINT            = "mds.dummy.address"
	KEY_SNAPSHOT_CLONE_DUMMY_ENDPOINT = "snapshot.clone.dummy.address"
	KEY_SNAPSHOT_CLONE_PROXY_ENDPOINT = "snapshot.clone.proxy.address"
	KEY_MONITOR_PROMETHEUS_ENDPOINT   = "monitor.prometheus.address"
)

type clusterConfig struct {
	Name   string `json:"name"`
	Config string `json:"config"`
}

type clusterServicesAddr struct {
	ClusterId int               `json:"clusterId"`
	Addrs     map[string]string `json:"addrs"`
}

func newAdmFail(r *pigeon.Request, err error) bool {
	r.Logger().Error("failed when new curveadm",
		pigeon.Field("error", err))
	return core.Exit(r, err)
}

func ListHost(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data, err := hosts.List(adm)
	if err != nil {
		r.Logger().Error("ListHost failed",
			pigeon.Field("error", err))
		return core.Exit(r, err)
	}
	return core.ExitSuccessWithData(r, data)
}

func CommitHost(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data := ctx.Data.(*CommitHostRequest)
	err = hosts.Commit(adm, data.Hosts)
	if err != nil {
		r.Logger().Error("CommitHosts failed",
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func ListDisk(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data, err := disks.List(adm)
	if err != nil {
		r.Logger().Error("ListDisks failed",
			pigeon.Field("error", err))
		return core.Exit(r, err)
	}
	return core.ExitSuccessWithData(r, data)
}

func CommitDisk(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data := ctx.Data.(*CommitDiskRequest)
	err = disks.Commit(adm, data.Disks)
	if err != nil {
		r.Logger().Error("CommitDisks failed",
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func GetFormatStatus(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data, err := command.Format(adm, true)
	if err != nil {
		r.Logger().Error("GetFormatStatus failed",
			pigeon.Field("error", err))
		return core.Exit(r, err)
	}
	return core.ExitSuccessWithData(r, data)
}

func FormatDisk(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	_, err = command.Format(adm, false)
	if err != nil {
		r.Logger().Error("FormatDisk failed",
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func ShowConfig(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	name, conf, err := config.Show(adm)
	if err != nil {
		r.Logger().Error("ShowClusterConfig failed",
			pigeon.Field("error", err))
		return core.Exit(r, err)
	}
	return core.ExitSuccessWithData(r, clusterConfig{Name: name, Config: conf})
}

func CommitConfig(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data := ctx.Data.(*CommitConfigRequest)
	err = config.Commit(adm, data.Name, data.Conf)
	if err != nil {
		r.Logger().Error("CommitConfig failed",
			pigeon.Field("cluster name", data.Name),
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func ListCluster(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	clusters, err := cluster.List(adm)
	if err != nil {
		r.Logger().Error("ListCluster failed",
			pigeon.Field("error", err))
		return core.Exit(r, err)
	}
	return core.ExitSuccessWithData(r, clusters)
}

func CheckoutCluster(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data := ctx.Data.(*CheckoutClusterRequest)
	err = cluster.Checkout(adm, data.Name)
	if err != nil {
		r.Logger().Error("Cluster checkout failed",
			pigeon.Field("name", data.Name),
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func AddCluster(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	data := ctx.Data.(*AddClusterRequest)
	err = cluster.Add(adm, data.Name, data.Desc, data.Topo)
	if err != nil {
		r.Logger().Error("AddCluster failed",
			pigeon.Field("name", data.Name),
			pigeon.Field("error", err))
		return core.Exit(r, err)
	}

	// checkout cluster
	err = cluster.Checkout(adm, data.Name)
	if err != nil {
		r.Logger().Error("Cluster checkout failed",
			pigeon.Field("name", data.Name),
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func DeployCluster(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}
	err = command.Deploy(adm)
	if err != nil {
		r.Logger().Error("DeployCluster failed",
			pigeon.Field("error", err))
	}
	return core.Exit(r, err)
}

func getServicesAddrFromConf(dcs []*topology.DeployConfig, mcs []*configure.MonitorConfig) map[string]string {
	etcdAddr := []string{}
	mdsAddr := []string{}
	mdsDummyAddr := []string{}
	snapshotCloneDummyAddr := []string{}
	snapshotCloneProxyAddr := []string{}
	var prometheusAddr string

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
	for _, mc := range mcs {
		if mc.GetRole() == configure.ROLE_GRAFANA {
			prometheusAddr = fmt.Sprintf("%s:%d", mc.GetPrometheusIp(), mc.GetPrometheusListenPort())
		}
	}
	ret := make(map[string]string)
	ret[KEY_ETCD_ENDPOINT] = strings.Join(etcdAddr, ",")
	ret[KEY_MDS_ENDPOINT] = strings.Join(mdsAddr, ",")
	ret[KEY_MDS_DUMMY_ENDPOINT] = strings.Join(mdsDummyAddr, ",")
	ret[KEY_SNAPSHOT_CLONE_DUMMY_ENDPOINT] = strings.Join(snapshotCloneDummyAddr, ",")
	ret[KEY_SNAPSHOT_CLONE_PROXY_ENDPOINT] = strings.Join(snapshotCloneProxyAddr, ",")
	ret[KEY_MONITOR_PROMETHEUS_ENDPOINT] = prometheusAddr
	return ret
}

func GetClusterServicesAddr(r *pigeon.Request, ctx *Context) bool {
	adm, err := cli.NewCurveAdm()
	if err != nil {
		return newAdmFail(r, err)
	}

	servicesAddr := clusterServicesAddr{}
	// check current cluster
	if adm.ClusterId() == -1 {
		r.Logger().Warn("GetClusterServicesAddr failed, no current cluster")
		return core.ExitSuccessWithData(r, servicesAddr)
	}

	// parse services addr
	hosts, hostIps, dcs, err := monitor.ParseTopology(adm)
	if err != nil {
		r.Logger().Warn("monitor.ParseTopology failed when GetClusterServicesAddr",
			pigeon.Field("error", err))
		return core.ExitSuccessWithData(r, servicesAddr)
	}

	monitor := adm.Monitor()
	mcs := []*configure.MonitorConfig{}
	if len(monitor.Monitor) != 0 {
		mcs, err = configure.ParseMonitorConfig(adm, "", monitor.Monitor, hosts, hostIps, dcs)
		if err != nil {
			r.Logger().Warn("ParseMonitorConfig failed when GetClusterServicesAddr",
				pigeon.Field("error", err))
			return core.ExitSuccessWithData(r, servicesAddr)
		}
	}
	servicesAddr.ClusterId = adm.ClusterId()
	servicesAddr.Addrs = getServicesAddrFromConf(dcs, mcs)
	return core.ExitSuccessWithData(r, servicesAddr)
}
