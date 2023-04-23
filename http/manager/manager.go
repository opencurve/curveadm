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
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/cli/command"
	"github.com/opencurve/curveadm/cli/command/cluster"
	"github.com/opencurve/curveadm/cli/command/config"
	"github.com/opencurve/curveadm/cli/command/disks"
	"github.com/opencurve/curveadm/cli/command/hosts"
	"github.com/opencurve/curveadm/http/core"
	"github.com/opencurve/pigeon"
)

type clusterConfig struct {
	Name   string `json:"name"`
	Config string `json:"config"`
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
