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
 * Created Date: 2022-01-14
 * Author: Jingli Chen (Wine93)
 */

package common

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/fatih/color"
)

const (
	PROMPT_REMOVE_CLUSTER = `{{.warning}}
Do you want to continue?`

	PROMPT_STOP_SERVICE = `{{.warning}}
Do you want to continue?`

	PROMPT_CLEAN_SERVICE = `{{.warning}}
  - Service role: {{.role}} ("*" means all roles)
  - Service host: {{.host}} ("*" means all hosts)
  - Clean items : [{{.items}}]
Do you want to continue?`

	PROMPT_COLLECT_SERVICE = `
FYI:
  > We have collected logs for troubleshooting,
  > and now we will send these logs to the curve center.
  > Please don't worry about the data security,
  > we guarantee that all logs are encrypted
  > and only you have the secret key.
`

	PROMPT_TOPOLOGY_CHANGE_NOTICE = `
NOTICE: We noticed that you have modified the configuration of 
some services while {{.operation}}. If you want make these 
configurations effect, you should reload the corresponding 
services after the scale out success.
`

	PROMPT_CANCEL_OPERATION = `[x] {{.operation}} canceled`

	DEFAULT_CONFIRM_PROMPT = "Do you want to continue?"
)

type Prompt struct {
	tmpl *template.Template
	data map[string]interface{}
}

func NewPrompt(text string) *Prompt {
	return &Prompt{
		tmpl: template.Must(template.New("prompt").Parse(text)),
		data: map[string]interface{}{},
	}
}

func (p *Prompt) Build() string {
	buffer := bytes.NewBufferString("")
	err := p.tmpl.Execute(buffer, p.data)
	if err != nil {
		return ""
	}
	return buffer.String()
}

func PromptRemoveCluster(clusterName string) string {
	prompt := NewPrompt(PROMPT_REMOVE_CLUSTER)
	prompt.data["warning"] = fmt.Sprintf("WARNING: cluster '%s' will be removed,\n"+
		"and all data in it will be cleaned up", clusterName)
	return prompt.Build()
}

func PromptStopService() string {
	prompt := NewPrompt(PROMPT_STOP_SERVICE)
	prompt.data["warning"] = "WARNING: stop service may cause client IO be hang"
	return prompt.Build()
}

func PromptCleanService(role, host string, items []string) string {
	prompt := NewPrompt(PROMPT_CLEAN_SERVICE)
	prompt.data["warning"] = "WARNING: service items which matched will be cleaned up"
	prompt.data["role"] = role
	prompt.data["host"] = host
	prompt.data["items"] = strings.Join(items, ",")
	return prompt.Build()
}

func PromptCollectService() string {
	prompt := NewPrompt(color.YellowString(PROMPT_COLLECT_SERVICE) + DEFAULT_CONFIRM_PROMPT)
	return prompt.Build()
}

func PromptScaleOut(warning bool) string {
	prompt := NewPrompt(DEFAULT_CONFIRM_PROMPT)
	if warning {
		prompt = NewPrompt(color.YellowString(PROMPT_TOPOLOGY_CHANGE_NOTICE) + DEFAULT_CONFIRM_PROMPT)
	}
	prompt.data["operation"] = "scale out cluster"
	return prompt.Build()
}

func PromptMigrate(warning bool) string {
	prompt := NewPrompt(DEFAULT_CONFIRM_PROMPT)
	if warning {
		prompt = NewPrompt(color.YellowString(PROMPT_TOPOLOGY_CHANGE_NOTICE) + DEFAULT_CONFIRM_PROMPT)
	}
	prompt.data["operation"] = "migrate services"
	return prompt.Build()
}

func PromptCancelOpetation(operation string) string {
	prompt := NewPrompt(color.RedString(PROMPT_CANCEL_OPERATION))
	prompt.data["operation"] = operation
	return prompt.Build()
}
