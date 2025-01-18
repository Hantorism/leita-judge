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
	language := "Java"
	problemId := "1000"
	testcases := 2
	command := Commands[language]

	fmt.Println("언어:", language)
	fmt.Println("문제 번호:", problemId)
	fmt.Println("빌드 명령어:", command.BuildCmd)
	fmt.Println("실행 명령어:", command.RunCmd)

	buildSource(command.BuildCmd)
	results := judge(command.RunCmd, problemId, testcases)
	report(results)
}

func buildSource(buildCmd []string) {
	fmt.Println("-----------------------")
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

func judge(runCmd []string, problemId string, testcases int) []bool {
	results := make([]bool, 0, testcases)

	for i := 0; i < testcases; i++ {
		fmt.Println("-----------------------")
		fmt.Printf("%d번째 테스트케이스 실행\n", i+1)

		inputFile := "problem/" + problemId + "/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := ioutil.ReadFile(inputFile)
		if err != nil {
			fmt.Printf("입력 파일 읽기 실패: %v\n", err)
			return nil
		}

		output, err := executeProgram(runCmd, inputContents)
		if err != nil {
			fmt.Printf("프로그램 실행 실패: %v\n", err)
			return nil
		}

		outputFile := "problem/" + problemId + "/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := ioutil.ReadFile(outputFile)
		if err != nil {
			fmt.Printf(".out 파일 읽기 실패: %v\n", err)
			return nil
		}

		result := checkDifference(output, outputContents)
		results = append(results, result)
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

func removeLineFeed(output []byte) []byte {
	if len(output) > 0 && output[len(output)-1] == 10 {
		return output[:len(output)-1]
	}
	return output
}

func executeProgram(runCmd []string, inputContents []byte) ([]byte, error) {
	fmt.Println("프로그램 실행 중...")
	cmd := exec.Command(runCmd[0], runCmd[1:]...)
	cmd.Stdin = bytes.NewReader(inputContents)

	outputBuffer := new(bytes.Buffer)
	cmd.Stdout = outputBuffer
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	output := outputBuffer.Bytes()
	output = removeLineFeed(output)
	return output, nil
}

func checkDifference(result, outputContents []byte) bool {
	fmt.Println("결과를 비교 중...")
	if !bytes.Equal(result, outputContents) {
		fmt.Println("결과가 일치하지 않습니다.")
		return false
	}
	fmt.Println("결과가 일치합니다!")
	return true
}
