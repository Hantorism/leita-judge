package utils

import (
	"os"

	"github.com/gofiber/fiber/v2/log"
)

func MakeDir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
