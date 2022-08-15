/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-07-14
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package checker

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	S3_TEMPLATE_VALUE = "<>"
)

type (
	// check whether host exist
	step2CheckSSHConfigure struct {
		dc       *topology.DeployConfig
		curveadm *cli.CurveAdm
	}

	// check whether the S3 configure is valid
	step2CheckS3Configure struct {
		dc       *topology.DeployConfig
		curveadm *cli.CurveAdm
	}

	// check whether directory path is absolute path
	step2CheckDirectoryPath struct {
		sequence int
		dc       *topology.DeployConfig
	}

	// check whether the data directory is duplicate
	step2CheckDataDirectoryDuplicate struct {
		dcs []*topology.DeployConfig
	}

	// check whether the listen port is duplicate
	step2CheckAddressDuplicate struct {
		dcs []*topology.DeployConfig
	}

	// check list:
	//   (1) each role requires at least 3 services
	//   (2) each requires at least 3 hosts
	step2CheckServices struct {
		curveadm *cli.CurveAdm
		dcs      []*topology.DeployConfig
	}
)

func (s *step2CheckSSHConfigure) Execute(ctx *context.Context) error {
	host := s.dc.GetHost()
	_, err := s.curveadm.GetHost(host)
	return err
}

func (s *step2CheckS3Configure) Execute(ctx *context.Context) error {
	dc := s.dc
	if s.curveadm.MemStorage().Get(comm.KEY_CHECK_SKIP_SNAPSHOECLONE).(bool) {
		return nil
	} else if dc.GetKind() != topology.KIND_CURVEBS {
		return nil
	}

	items := []struct {
		key   string
		value string
		err   *errno.ErrorCode
	}{
		{topology.CONFIG_S3_ACCESS_KEY.Key(), dc.GetS3AccessKey(), errno.ERR_INVALID_S3_ACCESS_KEY},
		{topology.CONFIG_S3_SECRET_KEY.Key(), dc.GetS3SecretKey(), errno.ERR_INVALID_S3_SECRET_KEY},
		{topology.CONFIG_S3_ADDRESS.Key(), dc.GetS3Address(), errno.ERR_INVALID_S3_ADDRESS},
		{topology.CONFIG_S3_BUCKET_NAME.Key(), dc.GetS3BucketName(), errno.ERR_INVALID_S3_BUCKET_NAME},
	}

	weak := s.curveadm.MemStorage().Get(comm.KEY_CHECK_WITH_WEAK).(bool) // for topology commit
	for _, item := range items {
		key := item.key
		value := item.value
		err := item.err
		if value == S3_TEMPLATE_VALUE {
			return err.F("%s: %s", key, value)
		} else if len(value) == 0 && !weak {
			return err.F("%s: %s", key, value)
		}
	}

	return nil
}

func (s *step2CheckDirectoryPath) Execute(ctx *context.Context) error {
	dc := s.dc
	dirs := getServiceDirectorys(dc)
	for _, dir := range dirs {
		if !strings.HasPrefix(dir.Path, "/") {
			return errno.ERR_DIRECTORY_REQUIRE_ABSOLUTE_PATH.
				F("%s.host[%s].%s: %s", dc.GetRole(), dc.GetHost(), dir.Type, dir.Path)
		}
	}
	return nil
}

func (s *step2CheckDataDirectoryDuplicate) Execute(ctx *context.Context) error {
	used := map[string]bool{}
	for _, dc := range s.dcs {
		dataDir := dc.GetDataDir()
		if len(dataDir) == 0 {
			continue
		}

		key := fmt.Sprintf("%s:%s", dc.GetHost(), dataDir)
		if _, ok := used[key]; ok {
			return errno.ERR_DATA_DIRECTORY_ALREADY_IN_USE.
				F("%s.host[%s].data_dir: %s", dc.GetRole(), dc.GetHost(), dataDir)
		}
		used[key] = true
	}
	return nil
}

func (s *step2CheckAddressDuplicate) Execute(ctx *context.Context) error {
	m := map[string]bool{}
	for _, dc := range s.dcs {
		addresses := getServiceListenAddresses(dc)
		for _, address := range addresses {
			addr := fmt.Sprintf("%s:%d", address.IP, address.Port)
			if _, ok := m[addr]; ok {
				return errno.ERR_DUPLICATE_LISTEN_ADDRESS.
					F("duplicate address: %s (%s.host[%s])", addr, dc.GetRole(), dc.GetHost())
			}
			m[addr] = true
		}
	}
	return nil
}

func (s *step2CheckServices) getHostNum(dcs []*topology.DeployConfig) int {
	num := 0
	exist := map[string]bool{}
	for _, dc := range dcs {
		parentId := dc.GetParentId()
		if _, ok := exist[parentId]; !ok {
			num++
			exist[parentId] = true
		}
	}
	return num
}

func (s *step2CheckServices) skip(role string) bool {
	kind := s.dcs[0].GetKind()
	// KIND_CURVEFS
	if kind == topology.KIND_CURVEFS {
		if role == ROLE_CHUNKSERVER || role == ROLE_SNAPSHOTCLONE {
			return true
		}
		return false
	}

	// KIND_CURVEBS
	skip := s.curveadm.MemStorage().Get(comm.KEY_CHECK_SKIP_SNAPSHOECLONE).(bool)
	if role == ROLE_METASERVER {
		return true
	} else if skip && role == ROLE_SNAPSHOTCLONE {
		return true
	}
	return false
}

func (s *step2CheckServices) Execute(ctx *context.Context) error {
	// (1): each role requires at least 3 services
	items := []struct {
		role string
		err  *errno.ErrorCode
	}{
		{ROLE_ETCD, errno.ERR_ETCD_REQUIRES_3_SERVICES},
		{ROLE_MDS, errno.ERR_MDS_REQUIRES_3_SERVICES},
		{ROLE_CHUNKSERVER, errno.ERR_CHUNKSERVER_REQUIRES_3_SERVICES},
		{ROLE_SNAPSHOTCLONE, errno.ERR_SNAPSHOTCLONE_REQUIRES_3_SERVICES}, // 0 OR >= 3
		{ROLE_METASERVER, errno.ERR_METASERVER_REQUIRES_3_SERVICES},
	}
	weak := s.curveadm.MemStorage().Get(comm.KEY_CHECK_WITH_WEAK).(bool) // for topology commit
	for _, item := range items {
		role := item.role
		if s.skip(role) {
			continue
		}
		dcs := s.curveadm.FilterDeployConfigByRole(s.dcs, role)
		num := len(dcs)
		if weak && role == ROLE_SNAPSHOTCLONE {
			if num == 0 || num >= 3 {
				continue
			}
			return item.err
		} else if num < 3 {
			return item.err
		}
	}

	// (2) each roles requires at least 3 hosts
	items = []struct {
		role string
		err  *errno.ErrorCode
	}{
		{ROLE_ETCD, errno.ERR_ETCD_REQUIRES_3_HOSTS},
		{ROLE_MDS, errno.ERR_MDS_REQUIRES_3_HOSTS},
		{ROLE_CHUNKSERVER, errno.ERR_CHUNKSERVER_REQUIRES_3_HOSTS},
		{ROLE_SNAPSHOTCLONE, errno.ERR_SNAPSHOTCLONE_REQUIRES_3_HOSTS}, // 0 OR >= 3
		{ROLE_METASERVER, errno.ERR_METASERVER_REQUIRES_3_HOSTS},
	}
	for _, item := range items {
		role := item.role
		if s.skip(role) {
			continue
		}
		dcs := s.curveadm.FilterDeployConfigByRole(s.dcs, role)
		num := s.getHostNum(dcs)
		if weak && role == ROLE_SNAPSHOTCLONE {
			if num == 0 || num >= 3 {
				continue
			}
			return item.err
		} else if num < 3 {
			return item.err
		}
	}

	return nil
}

func NewCheckTopologyTask(curveadm *cli.CurveAdm, null interface{}) (*task.Task, error) {
	// new task
	dcs := curveadm.MemStorage().Get(comm.KEY_ALL_DEPLOY_CONFIGS).([]*topology.DeployConfig)
	subname := fmt.Sprintf("cluster=%s kind=%s", curveadm.ClusterName(), dcs[0].GetKind())
	t := task.NewTask("Check Topology <topology>", subname, nil)

	// add step to task
	for _, dc := range dcs {
		t.AddStep(&step2CheckSSHConfigure{
			dc:       dc,
			curveadm: curveadm,
		})
	}
	for _, dc := range dcs {
		t.AddStep(&step2CheckDirectoryPath{dc: dc})
	}
	t.AddStep(&step2CheckDataDirectoryDuplicate{dcs: dcs})
	t.AddStep(&step2CheckAddressDuplicate{dcs: dcs})
	t.AddStep(&step2CheckServices{
		dcs:      dcs,
		curveadm: curveadm,
	})
	for _, dc := range dcs {
		t.AddStep(&step2CheckS3Configure{
			dc:       dc,
			curveadm: curveadm,
		})
	}

	return t, nil
}
