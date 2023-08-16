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
 * Created Date: 2022-07-22
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package glg

import (
	"fmt"
	"strconv"

	"github.com/kpango/glg"
)

const (
	INDENT_SPACE = 4
	SPACE        = " "
)

func convertLevel(level string) glg.LEVEL {
	switch level {
	case "debug":
		return glg.DEBG
	case "info":
		return glg.INFO
	case "warn":
		return glg.WARN
	case "error":
		return glg.ERR
	default:
		return glg.DEBG
	}
}

func Init(level, filename string) error {
	writer := glg.FileWriter(filename, 0666)

	glg.Get().
		SetMode(glg.WRITER). // default is STD
		SetLevel(convertLevel(level)).
		SetLineTraceMode(glg.TraceLineShort).
		AddLevelWriter(glg.DEBG, writer).
		AddLevelWriter(glg.INFO, writer).
		AddLevelWriter(glg.WARN, writer).
		AddLevelWriter(glg.ERR, writer)

	return nil
}

func Field(key string, val interface{}) string {
	switch val := val.(type) {
	case bool:
		return fmt.Sprintf("%s: %s", key, strconv.FormatBool(val))
	case string:
		return fmt.Sprintf("%s: %s", key, val)
	case []byte:
		return fmt.Sprintf("%s: %s", key, string(val))
	case int:
	case int64:
		return fmt.Sprintf("%s: %d", key, val)
	case error:
		return fmt.Sprintf("%s: %s", key, val.Error())
	}
	return fmt.Sprintf("%s: %v", key, val)
}

func format(message string, val ...string) string {
	output := message + "\n"
	for _, v := range val {
		output = output + fmt.Sprintf("%*s%s\n", INDENT_SPACE, SPACE, v)
	}

	return output
}

func Debug(message string, val ...string) error {
	return glg.Debug(format(message, val...))
}

func Info(message string, val ...string) error {
	return glg.Info(format(message, val...))
}

func Warn(message string, val ...string) error {
	return glg.Warn(format(message, val...))
}

func Error(message string, val ...string) error {
	return glg.Error(format(message, val...))
}

func SwitchLevel(err error) func(message string, val ...string) error {
	if err != nil {
		return Error
	}
	return Info
}
