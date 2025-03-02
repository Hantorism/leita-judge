package services

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	. "leita/src/entities"
	. "leita/src/functions"
	"leita/src/repositories"
	. "leita/src/utils"
)

type ProblemService interface {
	SubmitProblem(dto SubmitProblemDTO) (SubmitProblemResult, error)
}

type problemService struct {
	repository repositories.ProblemRepository
}

func NewProblemService() (ProblemService, error) {
	repository, err := repositories.NewProblemRepository()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &problemService{
		repository: repository,
	}, nil
}

func (service *problemService) SubmitProblem(dto SubmitProblemDTO) (SubmitProblemResult, error) {
	problemId := dto.ProblemId
	submitId := dto.SubmitId
	language := dto.Language
	code := dto.Code
	testcases := dto.Testcases
	buildCmd := dto.BuildCmd
	runCmd := dto.RunCmd
	deleteCmd := dto.DeleteCmd

	log.Info("언어:", language)
	log.Info("제출 번호:", submitId)
	log.Info("문제 번호:", problemId)
	log.Info("코드 길이:", len(string(code)))
	log.Info("제출 코드:")
	log.Info(string(code))

	if err := MakeDir("submit/" + strconv.Itoa(submitId) + "/"); err != nil {
		log.Fatal(err)
		return SubmitProblemResult{
			Status:       fiber.StatusInternalServerError,
			IsSuccessful: false,
			Error:        err,
		}, err
	}

	defer func() {
		saveSubmitResultDAO := SaveSubmitResultDAO{
			Result:     "CORRECT",
			UsedMemory: 1,
			UsedTime:   1,
			SubmitId:   submitId,
		}

		if err := service.repository.SaveSubmitResult(saveSubmitResultDAO); err != nil {
			log.Fatal(err)
		}
	}()

	if err := buildSource(submitId, language, code, buildCmd); err != nil {
		log.Fatal(err)
		return SubmitProblemResult{
			Status:       fiber.StatusInternalServerError,
			IsSuccessful: false,
			Error:        err,
		}, err
	}

	defer func(language string, deleteCmd []string) {
		err := deleteProgram(language, deleteCmd)
		if err != nil {
			log.Fatal(err)
		}
	}(language, deleteCmd)

	results, err := judge(runCmd, problemId, testcases)
	if err != nil {
		log.Fatal(err)
		return SubmitProblemResult{
			Status:       fiber.StatusInternalServerError,
			IsSuccessful: false,
			Error:        err,
		}, err
	}

	if judgeResult := report(results); judgeResult != JudgePass {
		return SubmitProblemResult{
			Status:       fiber.StatusOK,
			IsSuccessful: false,
			Error:        nil,
		}, nil
	}

	return SubmitProblemResult{
		Status:       fiber.StatusOK,
		IsSuccessful: true,
		Error:        nil,
	}, nil
}

func buildSource(submitId int, language string, code []byte, buildCmd []string) error {
	log.Info("-----------------------")
	log.Info("소스 파일 저장 중...")
	inputFile := "submit/" + strconv.Itoa(submitId) + "/Main." + FileExtension(language)
	if err := os.WriteFile(inputFile, code, 0644); err != nil {
		return err
	}
	log.Info("저장 완료!")

	log.Info("소스 파일 빌드 중...")
	if len(buildCmd) == 0 {
		log.Info(language + " 빌드 생략")
		return nil
	}

	cmd := exec.Command(buildCmd[0], buildCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
		return err
	}

	log.Info("빌드 완료!")
	return nil
}

func judge(runCmd []string, problemId, testcases int) ([]bool, error) {
	results := make([]bool, 0, testcases)

	for i := 0; i < testcases; i++ {
		log.Info("-----------------------")
		log.Info("%d번째 테스트케이스 실행\n", i+1)

		inputFile := "problem/" + strconv.Itoa(problemId) + "/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := os.ReadFile(inputFile)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		output, err := executeProgram(runCmd, inputContents)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		outputFile := "problem/" + strconv.Itoa(problemId) + "/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := os.ReadFile(outputFile)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		result := checkDifference(output, outputContents)
		results = append(results, result)
	}

	return results, nil
}

func report(results []bool) JudgeResultEnum {
	log.Info("-----------------------")
	if len(results) < 1 {
		log.Info(JudgeError.String())
		return JudgeError
	}

	if !All(results) {
		log.Info(JudgeFail.String())
		return JudgeFail
	}

	log.Info(JudgePass.String())
	return JudgePass
}

func removeLineFeed(output []byte) []byte {
	if len(output) > 0 && output[len(output)-1] == 10 {
		return output[:len(output)-1]
	}

	return output
}

func executeProgram(runCmd []string, inputContents []byte) ([]byte, error) {
	log.Info("프로그램 실행 중...")
	cmd := exec.Command(runCmd[0], runCmd[1:]...)
	cmd.Stdin = bytes.NewReader(inputContents)

	outputBuffer := new(bytes.Buffer)
	cmd.Stdout = outputBuffer
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
		return nil, err
	}

	output := outputBuffer.Bytes()
	output = removeLineFeed(output)
	return output, nil
}

func checkDifference(result, outputContents []byte) bool {
	log.Info("예상 결과")
	log.Info(result)
	log.Info(string(result))
	log.Info("실제 결과")
	log.Info(outputContents)
	log.Info(string(outputContents))

	log.Info("결과를 비교 중...")
	if !bytes.Equal(result, outputContents) {
		log.Info("결과가 일치하지 않습니다.")
		return false
	}

	log.Info("결과가 일치합니다!")
	return true
}

func deleteProgram(language string, deleteCmd []string) error {
	log.Info("-----------------------")
	log.Info("생성된 실행 파일 삭제 중...")

	if len(deleteCmd) == 0 {
		log.Info(language + " 삭제 생략")
		return nil
	}

	cmd := exec.Command(deleteCmd[0], deleteCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
		return err
	}

	log.Info("실행 파일 삭제 완료!")
	return nil
}
