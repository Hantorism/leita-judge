package utils

import (
	"os"

	"github.com/gofiber/fiber/v2/log"
)

func GetTestCaseNum(path string) (int, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return len(entries), nil
}
