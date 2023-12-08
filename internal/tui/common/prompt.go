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

// __SIGN_BY_WINE93__

package common

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/fatih/color"
)

const (
	PROMPT_WARNING = `{{.warning}}
`

	PROMPT_COMMON_WARNING = `{{.warning}} 
  - Service id: {{.id}} ("*" means all id)
  - Service role: {{.role}} ("*" means all roles)
  - Service host: {{.host}} ("*" means all hosts)
`

	PROMPT_CLEAN_SERVICE = `{{.warning}}
  - Service role: {{.role}} ("*" means all roles)
  - Service host: {{.host}} ("*" means all hosts)
  - Clean items : [{{.items}}]
`

	PROMPT_COLLECT_SERVICE = `FYI:
  > We will collect service logs for troubleshooting, 
  > and send these logs to the curve center.
  > Please don't worry about the data security,
  > we guarantee that all logs are encrypted
  > and only you have the secret key.
`

	PROMPT_TOPOLOGY_CHANGE_NOTICE = `
NOTICE: If you have modified the configuration of some services while 
{{.operation}} and you want make these configurations effect, you 
should reload the corresponding services after the {{.operation}} success.
`

	PROMPT_FORMAT = `
NOTICE: Now we run all formating container successfully and it will
format disk in the background, please make sure that the formatting 
all done before deploy cluster, you can use the "curveadm format --status" 
to watch the formatting progress.
`
	PROMPT_CANCEL_OPERATION = `[x] {{.operation}} canceled`

	PROMPT_PATH_EXIST = `{{.path}} already exists.
`

	DEFAULT_CONFIRM_PROMPT = "Do you want to continue?"
)

var (
	PROMPT_ERROR_CODE = strings.Join([]string{
		color.CyanString("---"),
		color.CyanString("Error-Code: ") + "{{.code}}",
		color.CyanString("Error-Description: ") + "{{.description}}",
		"{{- if .clue}}",
		color.CyanString("Error-Clue: ") + "{{.clue}}",
		"{{- end}}",
		color.CyanString("How to Solve:"),
		color.CyanString("  * Website: ") + "{{.website}}",
		"{{- if .logpath}}",
		color.CyanString("  * Log: ") + "{{.logpath}}",
		"{{- end}}",
		color.CyanString("  * WeChat: ") + "{{.wechat}}",
	}, "\n")

	PROMPT_AUTO_UPGRADE = strings.Join([]string{
		color.MagentaString("CurveAdm {{.version}} released, we recommend you to upgrade it."),
		"Upgrade curveadm to {{.version}}?",
	}, "\n")
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
	prompt := NewPrompt(color.YellowString(PROMPT_WARNING) + DEFAULT_CONFIRM_PROMPT)
	prompt.data["warning"] = fmt.Sprintf("WARNING: cluster '%s' will be removed,\n"+
		"and all data in it will be cleaned up", clusterName)
	return prompt.Build()
}

func PromptFormat() string {
	return color.YellowString(PROMPT_FORMAT)
}

func PromptScaleOut() string {
	prompt := NewPrompt(color.YellowString(PROMPT_TOPOLOGY_CHANGE_NOTICE) + DEFAULT_CONFIRM_PROMPT)
	prompt.data["operation"] = "scale out cluster"
	return prompt.Build()
}

func PromptMigrate() string {
	prompt := NewPrompt(color.YellowString(PROMPT_TOPOLOGY_CHANGE_NOTICE) + DEFAULT_CONFIRM_PROMPT)
	prompt.data["operation"] = "migrate services"
	return prompt.Build()
}

func PromptStartService(id, role, host string) string {
	prompt := NewPrompt(color.YellowString(PROMPT_COMMON_WARNING) + DEFAULT_CONFIRM_PROMPT)
	prompt.data["warning"] = "WARNING: service items which matched will start"
	prompt.data["id"] = id
	prompt.data["role"] = role
	prompt.data["host"] = host
	return prompt.Build()
}

func PromptStopService(id, role, host string) string {
	prompt := NewPrompt(color.YellowString(PROMPT_COMMON_WARNING) + DEFAULT_CONFIRM_PROMPT)
	prompt.data["warning"] = "WARNING: stop service may cause client IO be hang"
	prompt.data["id"] = id
	prompt.data["role"] = role
	prompt.data["host"] = host
	return prompt.Build()
}

func PromptRestartService(id, role, host string) string {
	prompt := NewPrompt(color.YellowString(PROMPT_COMMON_WARNING) + DEFAULT_CONFIRM_PROMPT)
	prompt.data["warning"] = "WARNING: service items which matched will restart"
	prompt.data["id"] = id
	prompt.data["role"] = role
	prompt.data["host"] = host
	return prompt.Build()
}

func PromptReloadService(id, role, host string) string {
	prompt := NewPrompt(color.YellowString(PROMPT_COMMON_WARNING) + DEFAULT_CONFIRM_PROMPT)
	prompt.data["warning"] = "WARNING: service items which matched will reload"
	prompt.data["id"] = id
	prompt.data["role"] = role
	prompt.data["host"] = host
	return prompt.Build()
}

func PromptCleanService(role, host string, items []string) string {
	prompt := NewPrompt(color.YellowString(PROMPT_CLEAN_SERVICE) + DEFAULT_CONFIRM_PROMPT)
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

func PromptIncrementFormat() string {
	prompt := NewPrompt(color.YellowString(PROMPT_WARNING) + DEFAULT_CONFIRM_PROMPT)
	prompt.data["warning"] = "WARNING: increment format will stop chunkserver service"
	return prompt.Build()
}

func prettyClue(clue string) string {
	items := strings.Split(clue, "\n")
	for {
		n := len(items)
		if len(items[n-1]) > 0 || n == 0 {
			break
		}
		items = items[:n-1]
	}
	sep := fmt.Sprintf("\n%s", strings.Repeat(" ", len("Error-Clue: ")))
	return strings.Join(items, sep)
}

func PromptErrorCode(code int, description, clue, logpath string) string {
	prompt := NewPrompt(color.CyanString(PROMPT_ERROR_CODE))
	prompt.data["code"] = fmt.Sprintf("%06d", code)
	prompt.data["description"] = description
	if len(clue) > 0 {
		prompt.data["clue"] = prettyClue(clue)
	}
	prompt.data["website"] = fmt.Sprintf("https://github.com/opencurve/curveadm/wiki/errno%d#%06d", code/100000, code)
	if len(logpath) > 0 {
		prompt.data["logpath"] = logpath
	}
	prompt.data["wechat"] = "opencurve_bot"

	return prompt.Build()
}

func PromptCancelOpetation(operation string) string {
	prompt := NewPrompt(color.YellowString(PROMPT_CANCEL_OPERATION))
	prompt.data["operation"] = operation
	return prompt.Build()
}

func PromptAutoUpgrade(version string) string {
	prompt := NewPrompt(PROMPT_AUTO_UPGRADE)
	prompt.data["version"] = version
	return prompt.Build()
}

func PromptPathExist(path string) string {
	prompt := NewPrompt(color.YellowString(PROMPT_PATH_EXIST) + DEFAULT_CONFIRM_PROMPT)
	prompt.data["path"] = path
	return prompt.Build()
}
