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
 * Created Date: 2022-09-08
 * Author: Jingli Chen (Wine93)
 */

package checker

import (
	"testing"

	"github.com/golang-module/carbon/v2"
	"github.com/stretchr/testify/assert"
)

func TestWaitNginxStart(t *testing.T) {
	assert := assert.New(t)

	seconds := []int{1, 2, 3, 10}
	for _, second := range seconds {
		start := carbon.Now().Timestamp()
		err := waitNginxStarted(second)(nil)
		end := carbon.Now().Timestamp()
		elapse := end - start
		assert.Nil(err)
		assert.GreaterOrEqual(elapse, second)
		assert.Less(elapse, second+1)
	}
}
