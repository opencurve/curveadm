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
 * Created Date: 2021-12-16
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

type VariantName struct {
	Name                string
	CompressName        string
	EncryptCompressName string
}

func RandFilename(dir string) string {
	return fmt.Sprintf("%s/%s", dir, RandString(8))
}

func NewVariantName(name string) VariantName {
	return VariantName{
		Name:                name,
		CompressName:        fmt.Sprintf("%s.tar.gz", name),
		EncryptCompressName: fmt.Sprintf("%s-encrypted.tar.gz", name),
	}
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func AbsPath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}

func GetFilePermissions(path string) int {
	info, err := os.Stat(path)
	if err != nil {
		return -1
	}

	return int(info.Mode())
}

func ReadFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func WriteFile(filename, data string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := file.WriteString(data)
	if err != nil {
		return err
	} else if n != len(data) {
		return fmt.Errorf("write abort")
	}

	return nil
}
