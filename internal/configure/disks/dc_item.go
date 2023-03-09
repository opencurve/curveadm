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
 * Project: CurveAdm
 * Created Date: 2023-02-24
 * Author: Lijin Xiong (lijin.xiong@zstack.io)
 */

package disks

import (
	"github.com/opencurve/curveadm/internal/common"
	comm "github.com/opencurve/curveadm/internal/configure/common"
)

const (
	DEFAULT_FORMAT_PERCENT = 90
)

var (
	itemset = comm.NewItemSet()

	CONFIG_GLOBAL_CONTAINER_IMAGE = itemset.Insert(
		common.DISK_FORMAT_CONTAINER_IMAGE,
		comm.REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_GLOBAL_FORMAT_PERCENT = itemset.Insert(
		common.DISK_FORMAT_PERCENT,
		comm.REQUIRE_POSITIVE_INTEGER,
		false,
		DEFAULT_FORMAT_PERCENT,
	)

	CONFIG_GLOBAL_SERVICE_MOUNT_DEVICE = itemset.Insert(
		common.DISK_SERVICE_MOUNT_DEVICE,
		comm.REQUIRE_BOOL,
		false,
		false,
	)

	CONFIG_GLOBAL_HOST = itemset.Insert(
		common.DISK_FILTER_HOST,
		comm.REQUIRE_STRING_SLICE,
		false,
		nil,
	)

	CONFIG_DISK_DEVICE = itemset.Insert(
		common.DISK_FILTER_DEVICE,
		comm.REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_DISK_MOUNT_POINT = itemset.Insert(
		common.DISK_FORMAT_MOUNT_POINT,
		comm.REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_DISK_HOST_EXCLUDE = itemset.Insert(
		common.DISK_EXCLUDE_HOST,
		comm.REQUIRE_STRING_SLICE,
		false,
		nil,
	)
)
