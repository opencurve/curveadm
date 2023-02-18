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

package common

import (
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	REQUIRE_ANY = iota
	REQUIRE_STRING
	REQUIRE_BOOL
	REQUIRE_INT
	REQUIRE_POSITIVE_INTEGER
	REQUIRE_SLICE
)

type (
	Item struct {
		key          string
		require      int
		exclude      bool        // exclude for service config
		defaultValue interface{} // nil means no default value
	}

	ItemSet struct {
		items    []*Item
		key2item map[string]*Item
	}
)

func (i *Item) Key() string {
	return i.key
}

func (i *Item) DefaultValue() interface{} {
	return i.defaultValue
}

func NewItemSet() *ItemSet {
	return &ItemSet{
		items:    []*Item{},
		key2item: map[string]*Item{},
	}
}

func (itemset *ItemSet) Insert(key string, require int, exclude bool, defaultValue interface{}) *Item {
	i := &Item{key, require, exclude, defaultValue}
	itemset.key2item[key] = i
	itemset.items = append(itemset.items, i)
	return i
}

func (itemset *ItemSet) Get(key string) *Item {
	return itemset.key2item[key]
}

func (itemset *ItemSet) GetAll() []*Item {
	return itemset.items
}

func (itemset *ItemSet) Build(key string, value interface{}) (interface{}, error) {
	item := itemset.Get(key)
	if item == nil {
		return value, nil
	}

	v, ok := utils.All2Str(value)
	if !ok {
		return nil, errno.ERR_UNSUPPORT_CONFIGURE_VALUE_TYPE.
			F("%s: %v", key, value)
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

	default:
		// do nothing
	}

	return value, nil
}
