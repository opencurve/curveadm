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

package disks

import (
	comm "github.com/opencurve/curveadm/internal/configure/common"
)

const (
	DEFAULT_FORMAT_PERCENT = 90
)

var (
	itemset = comm.NewItemSet()

	CONFIG_DEVICE = itemset.Insert(
		"device",
		comm.REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_MOUNT_POINT = itemset.Insert(
		"mount",
		comm.REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_FORMAT_PERCENT = itemset.Insert(
		"format_percent",
		comm.REQUIRE_POSITIVE_INTEGER,
		false,
		DEFAULT_FORMAT_PERCENT,
	)

	CONFIG_HOSTS_EXCLUDE = itemset.Insert(
		"hosts_exclude",
		comm.REQUIRE_SLICE,
		false,
		nil,
	)

	CONFIG_HOSTS_ONLY = itemset.Insert(
		"hosts_only",
		comm.REQUIRE_SLICE,
		false,
		nil,
	)

	CONFIG_CONTAINER_IMAGE = itemset.Insert(
		"container_image",
		comm.REQUIRE_STRING,
		false,
		nil,
	)
)
