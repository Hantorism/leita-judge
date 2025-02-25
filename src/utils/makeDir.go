package utils

import (
	"fmt"
	"os"
)

func MakeDir() {
	if err := os.MkdirAll("submit", os.ModePerm); err != nil {
		fmt.Printf("디렉토리 생성 실패: %v\n", err)
		return
	}
}
