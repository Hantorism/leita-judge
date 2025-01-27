package problem

import (
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"io/ioutil"
	. "leita/src/commands"
	. "leita/src/function"
	. "leita/src/models"
	. "leita/src/utils"
	"os"
	"os/exec"
	"strconv"
)

type JudgeRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

func JudgeProblem(c fiber.Ctx) error {
	var req JudgeRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	problemId := c.Params("problemId")
	language := req.Language
	code := Decode(req.Code)
	testcases := 2
	command := Commands[language]

	fmt.Println("언어:", language)
	fmt.Println("문제 번호:", problemId)
	fmt.Println("제출 코드:")
	fmt.Println(string(code))

	if err := buildSource(language, problemId, code, command.RequireBuild, command.BuildCmd); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"isSuccessful": false,
			"error":        err.Error(),
		})
	}

	defer deleteProgram(language, command.RequireBuild, command.DeleteCmd)

	results, err := judge(command.RunCmd, problemId, testcases)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"isSuccessful": false,
			"error":        err.Error(),
		})
	}

	if judgeResult := report(results); judgeResult != JudgePass {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"isSuccessful": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"isSuccessful": true,
	})
}

func buildSource(language, problemId string, code []byte, requireBuild bool, buildCmd []string) error {
	fmt.Println("-----------------------")
	fmt.Println("소스 파일 저장 중...")
	inputFile := "submit/temp/Main." + FileExtension(language)
	if err := WriteStringToFile(inputFile, string(code)); err != nil {
		return fmt.Errorf("파일 저장 실패: %v\n", err)
	}
	fmt.Println("저장 완료!")

	fmt.Println("소스 파일 빌드 중...")
	if !requireBuild {
		fmt.Println(language + " 빌드 생략")
		return nil
	}

	cmd := exec.Command(buildCmd[0], buildCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("소스 파일 빌드 실패: %v\n", err)
	}

	fmt.Println("빌드 완료!")
	return nil
}

func judge(runCmd []string, problemId string, testcases int) ([]bool, error) {
	results := make([]bool, 0, testcases)

	for i := 0; i < testcases; i++ {
		fmt.Println("-----------------------")
		fmt.Printf("%d번째 테스트케이스 실행\n", i+1)

		inputFile := "problem/" + problemId + "/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := ioutil.ReadFile(inputFile)
		if err != nil {
			return nil, fmt.Errorf("입력 파일 읽기 실패: %v\n", err)
		}

		output, err := executeProgram(runCmd, inputContents)
		if err != nil {
			return nil, fmt.Errorf("프로그램 실행 실패: %v\n", err)
		}

		outputFile := "problem/" + problemId + "/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := ioutil.ReadFile(outputFile)
		if err != nil {
			return nil, fmt.Errorf(".out 파일 읽기 실패: %v\n", err)
		}

		result := checkDifference(output, outputContents)
		results = append(results, result)
	}

	return results, nil
}

func report(results []bool) JudgeResult {
	fmt.Println("-----------------------")
	if len(results) < 1 {
		fmt.Println(JudgeError.String())
		return JudgeError
	}
	if !All(results) {
		fmt.Println(JudgeFail.String())
		return JudgeFail
	}
	fmt.Println(JudgePass.String())
	return JudgePass
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

	if err := cmd.Run(); err != nil {
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

func deleteProgram(language string, requireBuild bool, deleteCmd []string) {
	fmt.Println("-----------------------")
	fmt.Println("생성된 실행 파일 삭제 중...")

	if !requireBuild {
		fmt.Println(language + " 삭제 생략")
		return
	}

	cmd := exec.Command(deleteCmd[0], deleteCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("실행 파일 삭제 실패: %v\n", err)
		return
	}
	fmt.Println("실행 파일 삭제 완료!")
}
