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

// __SIGN_BY_WINE93__

package topology

import (
	"github.com/mitchellh/hashstructure/v2"
	"github.com/opencurve/curveadm/internal/errno"
)

const (
	DIFF_ADD    int = 0
	DIFF_DELETE int = 1
	DIFF_CHANGE int = 2
)

type TopologyDiff struct {
	DiffType     int
	DeployConfig *DeployConfig
}

func hash(dc *DeployConfig) (uint64, error) {
	return hashstructure.Hash(*dc, hashstructure.FormatV2, nil)
}

func same(dc1, dc2 *DeployConfig) (bool, error) {
	hash1, err := hash(dc1)
	if err != nil {
		return false, errno.ERR_CREATE_HASH_FOR_TOPOLOGY_FAILED.E(err)
	}

	hash2, err := hash(dc2)
	if err != nil {
		return false, errno.ERR_CREATE_HASH_FOR_TOPOLOGY_FAILED.E(err)
	}

	return hash1 == hash2, nil
}

// return ids which belong to ids1, but not belong to ids2
func difference(ids1, ids2 map[string]*DeployConfig) map[string]*DeployConfig {
	ids := map[string]*DeployConfig{}
	for k, v := range ids1 {
		if _, ok := ids2[k]; !ok {
			ids[k] = v
		}
	}

	return ids
}

func DiffTopology(data1, data2 string, ctx *Context) ([]TopologyDiff, error) {
	var dcs1, dcs2 []*DeployConfig
	var err error

	dcs1, err = ParseTopology(data1, ctx)
	if err != nil {
		return nil, err
	}

	dcs2, err = ParseTopology(data2, ctx)
	if err != nil {
		return nil, err
	}

	ids1 := map[string]*DeployConfig{}
	for _, dc := range dcs1 {
		ids1[dc.GetId()] = dc
	}

	ids2 := map[string]*DeployConfig{}
	for _, dc := range dcs2 {
		ids2[dc.GetId()] = dc
	}

	diffs := []TopologyDiff{}

	// DELETE
	deleteIds := difference(ids1, ids2)
	for _, dc := range deleteIds {
		diffs = append(diffs, TopologyDiff{
			DiffType:     DIFF_DELETE,
			DeployConfig: dc,
		})
	}

	// ADD
	addIds := difference(ids2, ids1)
	for _, dc := range addIds {
		diffs = append(diffs, TopologyDiff{
			DiffType:     DIFF_ADD,
			DeployConfig: dc,
		})
	}

	// CHANGE
	for id, dc := range ids2 {
		if _, ok := deleteIds[id]; ok {
			continue
		} else if _, ok := addIds[id]; ok {
			continue
		}

		ok, err := same(ids1[id], dc)
		if err != nil {
			return nil, err
		} else if !ok {
			diffs = append(diffs, TopologyDiff{
				DiffType:     DIFF_CHANGE,
				DeployConfig: dc,
			})
		}
	}

	return diffs, nil
}
