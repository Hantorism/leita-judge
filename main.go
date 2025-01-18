package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	. "leita/src/commands"
	. "leita/src/function"
	"os"
	"os/exec"
	"strconv"
)

func main() {
	language := "C"
	testcases := 2
	command := Commands[language]

	buildSource(command.BuildCmd)
	results := judge(command.RunCmd, testcases)
	report(results)
}

func buildSource(buildCmd []string) {
	fmt.Println("소스 파일을 빌드 중...")

	cmd := exec.Command(buildCmd[0], buildCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("소스 파일 빌드 실패: %v\n", err)
		return
	}
	fmt.Println("빌드 완료!")
}

func judge(runCmd []string, testcases int) []bool {
	results := make([]bool, 0, testcases)

	for i := 0; i < testcases; i++ {
		fmt.Println("-----------------------")
		fmt.Printf("%d번째 테스트케이스 실행\n", i+1)

		inputFile := "problem/1000/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := ioutil.ReadFile(inputFile)
		if err != nil {
			fmt.Printf("입력 파일 읽기 실패: %v\n", err)
			return nil
		}

		fmt.Println("프로그램 실행 중...")
		cmd := exec.Command(runCmd[0], runCmd[1:]...)
		cmd.Stdin = bytes.NewReader(inputContents)

		outputBuffer := new(bytes.Buffer)
		cmd.Stdout = outputBuffer
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			fmt.Printf("프로그램 실행 실패: %v\n", err)
			return nil
		}
		result := outputBuffer.Bytes()
		result = removeLineFeed(result)

		outputFile := "problem/1000/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := ioutil.ReadFile(outputFile)
		if err != nil {
			fmt.Printf(".out 파일 읽기 실패: %v\n", err)
			return nil
		}

		fmt.Println("결과를 비교 중...")
		if bytes.Equal(result, outputContents) {
			fmt.Println("결과가 일치합니다!")
			results = append(results, true)
		} else {
			fmt.Println("결과가 일치하지 않습니다.")
			results = append(results, false)
		}
	}

	return results
}

func report(results []bool) {
	fmt.Println("-----------------------")
	if !All(results) {
		fmt.Println("문제를 틀렸습니다.")
		return
	}
	fmt.Println("문제를 맞췄습니다!")
}

func removeLineFeed(result []byte) []byte {
	if len(result) > 0 && result[len(result)-1] == 10 {
		return result[:len(result)-1]
	}
	return result
}
