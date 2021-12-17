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
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
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

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
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
