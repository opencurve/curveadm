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

package variable

import (
	"fmt"
	"regexp"

	log "github.com/opencurve/curveadm/pkg/log/glg"
)

const (
	REGEX_VARIABLE = `\${([^${}]+)}` // ${var_name}
)

type Variable struct {
	Name        string
	Description string
	Value       string
	Resolved    bool
}

type Variables struct {
	m map[string]*Variable
	r *regexp.Regexp
}

func NewVariables() *Variables {
	return &Variables{
		m: map[string]*Variable{},
	}
}

func (vars *Variables) Register(v Variable) error {
	name := v.Name
	if _, ok := vars.m[name]; ok {
		return fmt.Errorf("variable '%s' duplicate define", name)
	}

	vars.m[name] = &v
	return nil
}

func (vars *Variables) Get(name string) (string, error) {
	v, ok := vars.m[name]
	if !ok {
		return "", fmt.Errorf("variable '%s' not found", name)
	} else if !v.Resolved {
		return "", fmt.Errorf("variable '%s' unresolved", name)
	}

	return v.Value, nil
}

func (vars *Variables) Set(name, value string) error {
	v, ok := vars.m[name]
	if !ok {
		return fmt.Errorf("variable '%s' not found", name)
	}

	v.Value = value
	v.Resolved = true
	return nil
}

func (vars *Variables) resolve(name string, marked map[string]bool) (string, error) {
	marked[name] = true
	v, ok := vars.m[name]
	if !ok {
		return "", fmt.Errorf("variable '%s' not defined", name)
	} else if v.Resolved {
		return v.Value, nil
	}

	matches := vars.r.FindAllStringSubmatch(v.Value, -1)
	if len(matches) == 0 { // no variable
		v.Resolved = true
		return v.Value, nil
	}

	// resolve all sub-variable
	for _, mu := range matches {
		name = mu[1]
		if _, err := vars.resolve(name, marked); err != nil {
			return "", err
		}
	}

	// ${var}
	v.Value = vars.r.ReplaceAllStringFunc(v.Value, func(name string) string {
		return vars.m[name[2:len(name)-1]].Value
	})
	v.Resolved = true
	return v.Value, nil
}

func (vars *Variables) Build() error {
	r, err := regexp.Compile(REGEX_VARIABLE)
	if err != nil {
		return err
	}

	vars.r = r
	for _, v := range vars.m {
		marked := map[string]bool{}
		if _, err := vars.resolve(v.Name, marked); err != nil {
			return err
		}
	}
	return nil
}

// "hello, ${varname}" => "hello, world"
func (vars *Variables) Rendering(s string) (string, error) {
	matches := vars.r.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 { // no variable
		return s, nil
	}

	var err error
	value := vars.r.ReplaceAllStringFunc(s, func(name string) string {
		val, e := vars.Get(name[2 : len(name)-1])
		if e != nil && err == nil {
			err = e
		}
		return val
	})
	return value, err
}

func (vars *Variables) Debug() {
	for _, v := range vars.m {
		err := log.Info("Variable", log.Field(v.Name, v.Value))
		if err != nil {
			return
		}
	}
}
