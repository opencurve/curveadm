package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

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

func (e PromptError) Error() string {
	errMessage := ""
	if e.Err != nil {
		errMessage = e.Err.Error()
	}
	return fmt.Sprintf("\n%s\n%s", errMessage, e.Prompt)
}

func ReadFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
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

func PathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TrimNewline(s string) string {
	return strings.TrimRight(s, "\r\n")
}

func Slice2Map(s []string) map[string]bool {
	m := map[string]bool{}
	for _, item := range s {
		m[item] = true
	}
	return m
}

func CheckAddrListen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	return err == nil && conn != nil
}
