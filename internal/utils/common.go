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
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	REGEX_IP = `^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4})`
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type PromptError struct {
	Err    error
	Prompt string
}

func Type(v interface{}) string {
	switch v.(type) {
	case bool:
		return "bool"
	case string:
		return "string"
	case int:
		return "int"
	case int64:
		return "int64"
	case map[string]interface{}:
		return "string_interface_map"
	case []interface{}:
		return "any_slice"
	case float64:
		return "float64"
	default:
		return "unknown"
	}
}

func IsBool(v interface{}) bool {
	return Type(v) == "bool"
}

func IsString(v interface{}) bool {
	return Type(v) == "string"
}

func IsInt(v interface{}) bool {
	return Type(v) == "int"
}

func IsInt64(v interface{}) bool {
	return Type(v) == "int64"
}

func IsStringAnyMap(v interface{}) bool {
	return Type(v) == "string_interface_map"
}

func IsAnySlice(v interface{}) bool {
	return Type(v) == "any_slice"
}

func IsFunc(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Func
}

func IsFloat64(v interface{}) bool {
	return Type(v) == "float64"
}

func All2Str(v interface{}) (value string, ok bool) {
	ok = true
	if IsString(v) {
		value = v.(string)
	} else if IsInt(v) {
		value = strconv.Itoa(v.(int))
	} else if IsBool(v) {
		value = strconv.FormatBool(v.(bool))
	} else {
		ok = false
	}
	return
}

// convert all to string
func Atoa(v interface{}) string {
	value, _ := All2Str(v)
	return value
}

func Str2Int(s string) (int, bool) {
	v, err := strconv.Atoi(s)
	return v, err == nil
}

func Str2Bool(s string) (bool, bool) { // value, ok
	v, err := strconv.ParseBool(s)
	return v, err == nil
}

func IsTrueStr(s string) bool {
	v, yes := Str2Bool(s)
	return yes && v
}

func TrimSuffixRepeat(s, suffix string) string {
	for {
		if !strings.HasSuffix(s, suffix) {
			break
		}
		s = strings.TrimSuffix(s, suffix)
	}
	return s
}

func Min(nums ...int) int {
	ret := nums[0]
	for _, num := range nums {
		if num < ret {
			ret = num
		}
	}
	return ret
}

func copy(src, dest map[string]interface{}) {
	for key, value := range src {
		switch src[key].(type) {
		case map[string]interface{}:
			dest[key] = map[string]interface{}{}
			copy(src[key].(map[string]interface{}), dest[key].(map[string]interface{}))
		default:
			dest[key] = value
		}
	}
}

func DeepCopy(src map[string]interface{}) map[string]interface{} {
	dest := map[string]interface{}{}
	copy(src, dest)
	return dest
}

func Choose(ok bool, first, second string) string {
	if ok {
		return first
	}
	return second
}

func (e PromptError) Error() string {
	errMessage := ""
	if e.Err != nil {
		errMessage = e.Err.Error()
	}
	return fmt.Sprintf("\n%s\n%s", errMessage, e.Prompt)
}

func MD5Sum(data string) string {
	m := md5.New()
	m.Write([]byte(data))
	return hex.EncodeToString(m.Sum(nil))
}

func Diff(s1 string, s2 string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(s1, s2, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	return dmp.DiffPrettyText(diffs)
}

func NewCommand(format string, a ...interface{}) *exec.Cmd {
	args := strings.Split(fmt.Sprintf(format, a...), " ")
	return exec.Command(args[0], args[1:]...)
}

func Slice2Map[T comparable](t []T) map[T]bool {
	m := map[T]bool{}
	for _, item := range t {
		m[item] = true
	}
	return m
}

func Locate(s []string) map[string]int {
	m := map[string]int{}
	for i, item := range s {
		m[item] = i
	}
	return m
}

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetCurrentUser() string {
	user, err := user.Current()
	if err != nil {
		return "root"
	}

	return user.Username
}

func GetCurrentHomeDir() string {
	user, err := user.Current()
	if err != nil {
		return "/root"
	}

	return user.HomeDir
}

func IsValidAddress(address string) bool {
	regex, err := regexp.Compile(REGEX_IP)
	if err != nil {
		return false
	}

	mu := regex.FindStringSubmatch(address)
	return len(mu) > 0
}

func ExecShell(format string, a ...interface{}) (string, error) {
	cmd := fmt.Sprintf(format, a...)
	bytes, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	return string(bytes), err
}

func EncryptFile(srcfile, dstfile, secret string) error {
	infile, err := os.Open(srcfile)
	if err != nil {
		return err
	}
	defer infile.Close()

	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return err
	}

	// Never use more than 2^32 random nonces with a given key
	// because of the risk of repeat.
	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(crand.Reader, iv); err != nil {
		return err
	}

	outfile, err := os.OpenFile(dstfile, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer outfile.Close()

	// The buffer size must be multiple of 16 bytes
	buf := make([]byte, 1024)
	stream := cipher.NewCTR(block, iv)
	for {
		n, err := infile.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			// Write into file
			outfile.Write(buf[:n])
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	// Append the IV
	outfile.Write(iv)
	return nil
}
