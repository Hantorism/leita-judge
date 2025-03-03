package utils

import (
	"encoding/base64"

	"github.com/gofiber/fiber/v2/log"
)

func Decode(encodedString string) ([]byte, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return decodedBytes, nil
}
