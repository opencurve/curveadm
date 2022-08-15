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
 * Created Date: 2022-01-07
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package tasks

import (
	"sync"

	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	STATUS_OK = iota
	STATUS_SKIP
	STATUS_ERROR
)

type monitor struct {
	err    error
	result map[int][]error // sub task result (key: progress bar id)
	mutex  sync.Mutex
}

func newMonitor() *monitor {
	return &monitor{
		err:    nil,
		result: map[int][]error{},
		mutex:  sync.Mutex{},
	}
}

func (m *monitor) error() error {
	return m.err
}

// return number of {success, skip, error}
func (m *monitor) sum(bid int) (int, int, int) {
	nsucc, nskip, nerr := 0, 0, 0
	for _, err := range m.result[bid] {
		if err == nil {
			nsucc++
		} else if err == task.ERR_SKIP_TASK {
			nskip++
		} else {
			nerr++
		}
	}
	return nsucc, nskip, nerr
}

func (m *monitor) set(bid int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.result[bid] = append(m.result[bid], err)
	if err != nil && err != task.ERR_SKIP_TASK {
		m.err = err
	}
}

func (m *monitor) get(bid int) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	nsucc, nskip, nerr := m.sum(bid)
	total := nsucc + nskip + nerr
	if nerr != 0 {
		return STATUS_ERROR
	} else if nskip == total {
		return STATUS_SKIP
	}
	// success all or part of skip
	return STATUS_OK
}
