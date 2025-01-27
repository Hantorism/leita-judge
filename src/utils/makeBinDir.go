package utils

import (
	"fmt"
	"os"
)

func MakeBinDir() {
	err := os.MkdirAll("./bin", os.ModePerm)
	if err != nil {
		fmt.Printf("디렉토리 생성 실패: %v\n", err)
		return
	}
}
