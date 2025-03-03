package utils

import (
	"io"
	"os"

	"github.com/gofiber/fiber/v2/log"
)

func CopyFile(srcFilePath, dstFilePath string) error {
	src, err := os.Open(srcFilePath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer func(src *os.File) {
		err = src.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(src)

	dst, err := os.Create(dstFilePath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer func(dst *os.File) {
		err = dst.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(dst)

	_, _ = io.Copy(dst, src)
	return nil
}
