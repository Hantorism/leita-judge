package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/gofiber/fiber/v2"
	. "leita/src/entities"
	. "leita/src/functions"
	"leita/src/repositories"
	. "leita/src/utils"
)

type ProblemService interface {
	JudgeProblem(dto JudgeProblemDTO) JudgeProblemResult
}

type problemService struct{}

func NewProblemService() ProblemService {
	return &problemService{}
}

func (service *problemService) JudgeProblem(dto JudgeProblemDTO) JudgeProblemResult {
	problemId := dto.ProblemId
	submitId := dto.SubmitId
	language := dto.Language
	code := dto.Code
	testcases := dto.Testcases
	requireBuild := dto.RequireBuild
	buildCmd := dto.BuildCmd
	runCmd := dto.RunCmd
	deleteCmd := dto.DeleteCmd

	fmt.Println("언어:", language)
	fmt.Println("제출 번호:", submitId)
	fmt.Println("문제 번호:", problemId)
	fmt.Println("코드 길이:", len(string(code)))
	fmt.Println("제출 코드:")
	fmt.Println(string(code))

	MakeDir("submit/" + strconv.Itoa(submitId) + "/")

	defer func() {
		saveJudgeResultDAO := SaveJudgeResultDAO{
			SubmitId:     submitId,
			ProblemId:    problemId,
			Result:       "CORRECT",
			SizeOfCode:   len(string(code)),
			UserId:       3,
			UsedLanguage: language,
			UsedMemory:   1,
			UsedTime:     1,
		}

		problemRepository := repositories.NewProblemRepository()
		if err := problemRepository.SaveJudgeResult(saveJudgeResultDAO); err != nil {
			fmt.Println(err)
		}
	}()

	if err := buildSource(submitId, language, code, requireBuild, buildCmd); err != nil {
		return JudgeProblemResult{
			Status:       fiber.StatusInternalServerError,
			IsSuccessful: false,
			Error:        err,
		}
	}

	defer deleteProgram(language, requireBuild, deleteCmd)

	results, err := judge(runCmd, problemId, testcases)
	if err != nil {
		return JudgeProblemResult{
			Status:       fiber.StatusInternalServerError,
			IsSuccessful: false,
			Error:        err,
		}
	}

	if judgeResult := report(results); judgeResult != JudgePass {
		return JudgeProblemResult{
			Status:       fiber.StatusOK,
			IsSuccessful: false,
			Error:        nil,
		}
	}

	return JudgeProblemResult{
		Status:       fiber.StatusOK,
		IsSuccessful: true,
		Error:        nil,
	}
}

func buildSource(submitId int, language string, code []byte, requireBuild bool, buildCmd []string) error {
	fmt.Println("-----------------------")
	fmt.Println("소스 파일 저장 중...")
	inputFile := "submit/" + strconv.Itoa(submitId) + "/Main." + FileExtension(language)
	if err := os.WriteFile(inputFile, code, 0644); err != nil {
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

func judge(runCmd []string, problemId, testcases int) ([]bool, error) {
	results := make([]bool, 0, testcases)

	for i := 0; i < testcases; i++ {
		fmt.Println("-----------------------")
		fmt.Printf("%d번째 테스트케이스 실행\n", i+1)

		inputFile := "problem/" + strconv.Itoa(problemId) + "/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := os.ReadFile(inputFile)
		if err != nil {
			return nil, fmt.Errorf("입력 파일 읽기 실패: %v\n", err)
		}

		output, err := executeProgram(runCmd, inputContents)
		if err != nil {
			return nil, fmt.Errorf("프로그램 실행 실패: %v\n", err)
		}

		outputFile := "problem/" + strconv.Itoa(problemId) + "/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := os.ReadFile(outputFile)
		if err != nil {
			return nil, fmt.Errorf(".out 파일 읽기 실패: %v\n", err)
		}

		result := checkDifference(output, outputContents)
		results = append(results, result)
	}

	return results, nil
}

func report(results []bool) JudgeResultEnum {
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
	fmt.Println("예상 결과")
	fmt.Println(result)
	fmt.Println(string(result))
	fmt.Println("실제 결과")
	fmt.Println(outputContents)
	fmt.Println(string(outputContents))

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
