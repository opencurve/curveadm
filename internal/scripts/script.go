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
 * Created Date: 2021-11-25
 * Author: Jingli Chen (Wine93)
 */

package scripts

import (
	"fmt"
	"os"
)

var scripts = map[string]string{}

func init() {
	scripts = map[string]string{
		"wait": WAIT,
	}
}

func Get(name string) (string, bool) {
	v, ok := scripts[name]
	return v, ok
}

func MountScript(name, path string) error {
	if v, ok := scripts[name]; !ok {
		return fmt.Errorf("script '%s' not found", name)
	} else if file, err := os.Open(name); err != nil {
		return err
	} else if n, err := file.WriteString(v); err != nil {
		return err
	} else if n != len(v) {
		return fmt.Errorf("write abort")
	}
	return nil
}
