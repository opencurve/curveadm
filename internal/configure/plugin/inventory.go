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
 * Created Date: 2022-05-18
 * Author: Jingli Chen (Wine93)
 */

package plugin

import (
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

type (
	Target struct {
		Name           string `mapstructure:"name"`
		User           string `mapstructure:"ssh_user"`
		Host           string `mapstructure:"ssh_host"`
		Port           uint `mapstructure:"ssh_port"`
		PrivateKeyFile string `mapstructure:"ssh_private_key"`
	}

	Inventory struct {
		Targets []Target `mapstructure:"servers"`
	}
)

func ParseInventory(filename string) ([]Target, error) {
	if !utils.PathExist(filename) {
		return nil, nil
	}

	parser := viper.New()
	parser.SetConfigFile(filename)
	parser.SetConfigType("yaml")
	err := parser.ReadInConfig()
	if err != nil {
		return nil, err
	}

	inventory := &Inventory{}
	err = parser.Unmarshal(inventory)
	if err != nil {
		return nil, err
	}

	return inventory.Targets, nil
}
