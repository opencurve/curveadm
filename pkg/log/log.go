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

package log

import (
	zaplog "github.com/pingcap/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Debug = zaplog.Debug
	Info  = zaplog.Info
	Warn  = zaplog.Warn
	Error = zaplog.Error
)

func Init(level, filename string) error {
	gl, props, err := zaplog.InitLogger(&zaplog.Config{
		Level:            level,
		File:             zaplog.FileLogConfig{Filename: filename},
		Format:           "text",
		DisableTimestamp: false,
	}, zap.AddStacktrace(zapcore.FatalLevel))

	if err != nil {
		return err
	}

	zaplog.ReplaceGlobals(gl, props)
	return nil
}

func SwitchLevel(err error) func(msg string, fields ...zap.Field) {
	if err != nil {
		return zaplog.Error
	}
	return zaplog.Info
}

func Field(key string, val interface{}) zap.Field {
	switch val := val.(type) {
	case bool:
		return zap.Bool(key, val)
	case string:
		return zap.String(key, val)
	case []byte:
		return zap.String(key, string(val))
	case int:
		return zap.Int(key, val)
	case int64:
		return zap.Int64(key, val)
	case error:
		return zap.String(key, val.Error())
	}
	return zap.Skip()
}
