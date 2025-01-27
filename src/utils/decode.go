package utils

import (
	"encoding/base64"
	"fmt"
)

func Decode(encodedString string) []byte {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		fmt.Println("디코딩 실패: ", err)
		return nil
	}

	return decodedBytes
}
