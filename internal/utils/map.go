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

package utils

import (
	"sync"
)

type SafeMap struct {
	sync.RWMutex
	Map         map[string]interface{}
	transaction bool
}

func NewSafeMap() *SafeMap {
	return &SafeMap{Map: map[string]interface{}{}}
}

func (m *SafeMap) Get(key string) interface{} {
	if m.transaction {
		val := m.Map[key]
		return val
	}
	m.RLock()
	defer m.RUnlock()
	val := m.Map[key]
	return val
}

func (m *SafeMap) Set(key string, value interface{}) {
	if m.transaction {
		m.Map[key] = value
		return
	}
	m.Lock()
	defer m.Unlock()
	m.Map[key] = value
}

func (m *SafeMap) TX(callback func(m *SafeMap) error) error {
	m.Lock()
	m.transaction = true
	err := callback(m)
	m.transaction = false
	m.Unlock()
	return err
}
