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
 * Created Date: 2022-08-01
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package hosts

import (
	"fmt"
	"github.com/opencurve/curveadm/internal/errno"

	"github.com/opencurve/curveadm/internal/utils"
)

const (
	REQUIRE_ANY = iota
	REQUIRE_INT
	REQUIRE_STRING
	REQUIRE_BOOL
	REQUIRE_POSITIVE_INTEGER
	REQUIRE_STRING_SLICE

	DEFAULT_SSH_PORT = 22
)

type (
	// config item
	item struct {
		key          string
		require      int
		exclude      bool        // exclude for service config
		defaultValue interface{} // nil means no default value
	}

	itemSet struct {
		items    []*item
		key2item map[string]*item
	}
)

var (
	itemset = &itemSet{
		items:    []*item{},
		key2item: map[string]*item{},
	}

	CONFIG_HOST = itemset.insert(
		"host",
		REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_HOSTNAME = itemset.insert(
		"hostname",
		REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_SSH_HOSTNAME = itemset.insert(
		"ssh_hostname",
		REQUIRE_STRING,
		false,
		nil,
	)

	CONFIG_USER = itemset.insert(
		"user",
		REQUIRE_STRING,
		false,
		func(hc *HostConfig) interface{} {
			return utils.GetCurrentUser()
		},
	)

	CONFIG_SSH_PORT = itemset.insert(
		"ssh_port",
		REQUIRE_POSITIVE_INTEGER,
		false,
		DEFAULT_SSH_PORT,
	)

	CONFIG_PRIVATE_CONFIG_FILE = itemset.insert(
		"private_key_file",
		REQUIRE_STRING,
		false,
		func(hc *HostConfig) interface{} {
			return fmt.Sprintf("%s/.ssh/id_rsa", utils.GetCurrentHomeDir())
		},
	)

	CONFIG_FORWARD_AGENT = itemset.insert(
		"forward_agent",
		REQUIRE_BOOL,
		false,
		false,
	)

	CONFIG_BECOME_USER = itemset.insert(
		"become_user",
		REQUIRE_STRING,
		false,
		nil,
	)
)

func convertSlice[T int | string | any](key, value any) ([]T, error) {
	var slice []T
	if !utils.IsAnySlice(value) || len(value.([]any)) == 0 {
		return slice, errno.ERR_CONFIGURE_VALUE_REQUIRES_NONEMPTY_SLICE
	}
	anySlice := value.([]any)
	switch anySlice[0].(type) {
	case T:
		for _, str := range anySlice {
			slice = append(slice, str.(T))
		}
	default:
		return slice, errno.ERR_UNSUPPORT_CONFIGURE_VALUE_TYPE.
			F("%s: %v", key, value)
	}

	return slice, nil
}

func (i *item) Key() string {
	return i.key
}

func (itemset *itemSet) insert(key string, require int, exclude bool, defaultValue interface{}) *item {
	i := &item{key, require, exclude, defaultValue}
	itemset.key2item[key] = i
	itemset.items = append(itemset.items, i)
	return i
}

func (itemset *itemSet) get(key string) *item {
	return itemset.key2item[key]
}

func (itemset *itemSet) getAll() []*item {
	return itemset.items
}

func (itemset *itemSet) Build(key string, value interface{}) (interface{}, error) {
	item := itemset.get(key)
	if item == nil {
		return value, nil
	}

	v, ok := utils.All2Str(value)
	if !ok {
		if !utils.IsAnySlice(value) {
			return nil, errno.ERR_UNSUPPORT_CONFIGURE_VALUE_TYPE.
				F("%s: %v", key, value)
		}
	}

	switch item.require {
	case REQUIRE_ANY:
		// do nothing

	case REQUIRE_STRING:
		if len(v) == 0 {
			return nil, errno.ERR_CONFIGURE_VALUE_REQUIRES_NON_EMPTY_STRING.
				F("%s: %v", key, value)
		} else {
			return v, nil
		}

	case REQUIRE_INT:
		if v, ok := utils.Str2Int(v); !ok {
			return nil, errno.ERR_CONFIGURE_VALUE_REQUIRES_INTEGER.
				F("%s: %v", key, value)
		} else {
			return v, nil
		}

	case REQUIRE_POSITIVE_INTEGER:
		if v, ok := utils.Str2Int(v); !ok {
			return nil, errno.ERR_CONFIGURE_VALUE_REQUIRES_INTEGER.
				F("%s: %v", key, value)
		} else if v <= 0 {
			return nil, errno.ERR_CONFIGURE_VALUE_REQUIRES_POSITIVE_INTEGER.
				F("%s: %v", key, value)
		} else {
			return v, nil
		}

	case REQUIRE_BOOL:
		if v, ok := utils.Str2Bool(v); !ok {
			return nil, errno.ERR_CONFIGURE_VALUE_REQUIRES_BOOL.
				F("%s: %v", key, value)
		} else {
			return v, nil
		}

	case REQUIRE_STRING_SLICE:
		return convertSlice[string](key, value)

	default:
		// do nothing
	}

	return value, nil
}
