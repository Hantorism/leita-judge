package utils

import (
	"encoding/base64"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
)

var envMap = make(map[string]string)

func Decode(encodedString string) ([]byte, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return decodedBytes, nil
}

func FileExtension(language string) string {
	switch language {
	case "C":
		return "c"
	case "CPP":
		return "cpp"
	case "GO":
		return "go"
	case "JAVA":
		return "java"
	case "JAVASCRIPT":
		return "js"
	case "KOTLIN":
		return "kt"
	case "PYTHON":
		return "py"
	case "SWIFT":
		return "swift"
	default:
		return "error"
	}
}

func GetTestCaseNum(path string) (int, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return len(entries), nil
}

func LoadEnv() error {
	if err := godotenv.Load(".env"); err != nil {
		log.Error(err)
		return err
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	return nil
}

func GetEnv(key string) string {
	if value, exists := envMap[key]; exists {
		return value
	}

	return ""
}

func MakeDir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func ReplaceCommand(args []string, judgeType string, submitID int) []string {
	replaced := make([]string, len(args))
	for i, arg := range args {
		arg = strings.ReplaceAll(arg, "{JUDGE_TYPE}", judgeType)
		replaced[i] = strings.ReplaceAll(arg, "{SUBMIT_ID}", strconv.Itoa(submitID))
	}
	return replaced
}
