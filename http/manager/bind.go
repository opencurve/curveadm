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

import "github.com/opencurve/pigeon"

var METHOD_REQUEST map[string]Request

type (
	HandlerFunc func(r *pigeon.Request, ctx *Context) bool

	Context struct {
		Data interface{}
	}

	Request struct {
		httpMethod string
		method     string
		vType      interface{}
		handler    HandlerFunc
	}
)

func init() {
	METHOD_REQUEST = map[string]Request{}
	for _, request := range requests {
		METHOD_REQUEST[request.method] = request
	}
}

type ListHostRequest struct{}

type CommitHostRequest struct {
	Hosts string `json:"hosts" binding:"required"`
}

type ListDiskRequest struct{}

type CommitDiskRequest struct {
	Disks string `json:"disks" binding:"required"`
}

type GetFormatStatusRequest struct{}

type FormatDiskRequest struct{}

type ShowConfigRequest struct{}

type CommitConfigRequest struct {
	Name string `json:"name" binding:"required"`
	Conf string `json:"conf" binding:"required"`
}

type ListClusterRequest struct{}

type CheckoutClusterRequest struct {
	Name string `json:"name" binding:"required"`
}

type AddClusterRequest struct {
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc"`
	Topo string `json:"topo"`
}

type DeployClusterRequest struct{}

type GetClusterServicesAddrRequest struct{}

var requests = []Request{
	{
		"GET",
		"host.list",
		ListHostRequest{},
		ListHost,
	},
	{
		"POST",
		"host.commit",
		CommitHostRequest{},
		CommitHost,
	},
	{
		"GET",
		"disk.list",
		ListDiskRequest{},
		ListDisk,
	},
	{
		"POST",
		"disk.commit",
		CommitDiskRequest{},
		CommitDisk,
	},
	{
		"GET",
		"disk.format.status",
		GetFormatStatusRequest{},
		GetFormatStatus,
	},
	{
		"GET",
		"disk.format",
		FormatDiskRequest{},
		FormatDisk,
	},
	{
		"GET",
		"config.show",
		ShowConfigRequest{},
		ShowConfig,
	},
	{
		"POST",
		"config.commit",
		CommitConfigRequest{},
		CommitConfig,
	},
	{
		"GET",
		"cluster.list",
		ListClusterRequest{},
		ListCluster,
	},
	{
		"POST",
		"cluster.add",
		AddClusterRequest{},
		AddCluster,
	},
	{
		"POST",
		"cluster.checkout",
		CheckoutClusterRequest{},
		CheckoutCluster,
	},
	{
		"GET",
		"cluster.deploy",
		DeployClusterRequest{},
		DeployCluster,
	},
	{
		"GET",
		"cluster.service.addr",
		GetClusterServicesAddrRequest{},
		GetClusterServicesAddr,
	},
}
