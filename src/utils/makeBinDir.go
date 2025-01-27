package utils

import (
	"fmt"
	"os"
)

func MakeBinDir() {
	if err := os.MkdirAll("./bin", os.ModePerm); err != nil {
		fmt.Printf("디렉토리 생성 실패: %v\n", err)
		return
	}
}
