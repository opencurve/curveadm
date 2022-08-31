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
 * Created Date: 2022-08-31
 * Author: Jingli Chen (Wine93)
 */

package checker

import (
	"testing"

	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/stretchr/testify/assert"
)

func TestCheckKernelVersion(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		version string
		err     error
	}{
		{"5.12.0", nil},
		{"4.9.65-netease", nil},
		{"4.19.0-16-amd64", nil},
		{"3.15.0", nil},
		{"3.15.0.0.", errno.ERR_UNRECOGNIZED_KERNEL_VERSION},
		{"3.14.9", errno.ERR_RENAMEAT_NOT_SUPPORTED_IN_CURRENT_KERNEL},
	}

	dc := topology.DeployConfig{}
	for _, t := range tests {
		lambda := checkKernelVersion(&t.version, &dc)
		assert.Equal(lambda(nil), t.err)
	}
}
