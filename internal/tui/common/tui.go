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

package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/opencurve/curveadm/internal/utils"
)

type DecorateMessage struct {
	Message  string
	Decorate func(string) string
}

func originMessage(message interface{}) string {
	if utils.IsString(message) {
		return message.(string)
	}
	return message.(DecorateMessage).Message
}

func realMessage(message interface{}) string {
	if utils.IsString(message) {
		return message.(string)
	}

	decorateMessage := message.(DecorateMessage)
	return decorateMessage.Decorate(decorateMessage.Message)
}

func getLength(lines [][]interface{}) (int, int) {
	if len(lines) == 0 {
		return 0, 0
	}
	return len(lines), len(lines[0])
}

func fixedLength(lines [][]interface{}) []int {
	fixed := []int{}
	n, m := getLength(lines)
	for j := 0; j < m; j++ {
		maxLen := 0
		for i := 0; i < n; i++ {
			message := originMessage(lines[i][j])
			if len(message) > maxLen {
				maxLen = len(message)
			}
		}
		fixed = append(fixed, maxLen)
	}
	return fixed
}

func FixedFormat(lines [][]interface{}, nspace int) string {
	output := ""
	spacing := strings.Repeat(" ", nspace)
	n, m := getLength(lines)
	fixed := fixedLength(lines)

	for i := 0; i < n; i++ {
		first := true
		for j := 0; j < m; j++ {
			if fixed[j] == 0 { // maybe cutted
				continue
			}

			if !first {
				output += spacing
			} else {
				first = false
			}

			padding := strings.Repeat(" ", fixed[j]-len(originMessage(lines[i][j])))
			output += realMessage(lines[i][j]) + padding
		}
		output += "\n"
	}

	return output
}

func FormatTitle(title []string) ([]interface{}, []interface{}) {
	first := []interface{}{}
	second := []interface{}{}
	for _, item := range title {
		first = append(first, item)
		second = append(second, strings.Repeat("-", len(item)))
	}
	return first, second
}

func CutColumn(lines [][]interface{}, column int) {
	for _, line := range lines {
		line[column] = ""
	}
}

func TrimContainerId(containerId string) string {
	containerId = strings.TrimRight(containerId, "\r\n")
	if len(containerId) <= 12 {
		return containerId
	}
	return containerId[:12]
}

func prompt(prompt string) string {
	if prompt != "" {
		prompt += " "
	}
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(input, "\n")
}

func ConfirmNo(format string, a ...interface{}) bool {
	ans := prompt(fmt.Sprintf(format, a...) + "(default=Y)")
	switch strings.TrimSpace(strings.ToLower(ans)) {
	case "n", "no":
		return true
	default:
		return false
	}
}

func ConfirmYes(format string, a ...interface{}) bool {
	ans := prompt(fmt.Sprintf(format, a...) + " [yes/no]: (default=no)")
	switch strings.TrimSpace(ans) {
	case "yes":
		return true
	default:
		return false
	}
}
