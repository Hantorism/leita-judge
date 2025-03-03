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
	RunProblem(dto RunProblemDTO) (RunProblemResult, error)
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
	buildCmd := dto.BuildCmd
	runCmd := dto.RunCmd
	deleteCmd := dto.DeleteCmd

	printSubmitProblemInfo(language, submitId, problemId, code)

	if err := copyTestCases(submitId, problemId); err != nil {
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

	if err := buildSource(submitId, language, "submit", code, buildCmd); err != nil {
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

	judgeResults, err := judge(runCmd, submitId, "submit")
	if err != nil {
		log.Fatal(err)
		return SubmitProblemResult{
			Status:       fiber.StatusInternalServerError,
			IsSuccessful: false,
			Error:        err,
		}, err
	}

	if judgeResult := report(judgeResults); judgeResult != JudgePass {
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

func (service *problemService) RunProblem(dto RunProblemDTO) (RunProblemResult, error) {
	problemId := dto.ProblemId
	submitId := dto.SubmitId
	language := dto.Language
	code := dto.Code
	testCases := dto.TestCases
	buildCmd := dto.BuildCmd
	runCmd := dto.RunCmd
	deleteCmd := dto.DeleteCmd

	if err := printRunProblemInfo(language, submitId, problemId, code, testCases); err != nil {
		log.Fatal(err)
		return RunProblemResult{
			Status:       fiber.StatusInternalServerError,
			IsSuccessful: false,
			Error:        err,
		}, err
	}

	if err := saveTestCases(submitId, testCases); err != nil {
		log.Fatal(err)
		return RunProblemResult{
			Status:       fiber.StatusInternalServerError,
			IsSuccessful: false,
			Error:        err,
		}, err
	}

	if err := buildSource(submitId, language, "run", code, buildCmd); err != nil {
		log.Fatal(err)
		return RunProblemResult{
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

	judgeResults, err := judge(runCmd, submitId, "run")
	if err != nil {
		log.Fatal(err)
		return RunProblemResult{
			Status:       fiber.StatusInternalServerError,
			IsSuccessful: false,
			Error:        err,
		}, err
	}

	if judgeResult := report(judgeResults); judgeResult != JudgePass {
		return RunProblemResult{
			Status:       fiber.StatusOK,
			IsSuccessful: false,
			Error:        nil,
		}, nil
	}

	return RunProblemResult{
		Status:       fiber.StatusOK,
		IsSuccessful: true,
		Error:        nil,
	}, nil
}

func printSubmitProblemInfo(language string, submitId int, problemId int, code []byte) {
	log.Info("언어: ", language)
	log.Info("제출 번호: ", submitId)
	log.Info("문제 번호: ", problemId)
	log.Info("코드 길이: ", len(string(code)))
	log.Info("제출 코드:\n", string(code))
}

func printRunProblemInfo(language string, submitId int, problemId int, code []byte, testCases []TestCase) error {
	log.Info("언어: ", language)
	log.Info("제출 번호: ", submitId)
	log.Info("문제 번호: ", problemId)
	log.Info("코드 길이: ", len(string(code)))
	log.Info("제출 코드:\n", string(code))
	log.Info("테스트 케이스:")
	for i, testCase := range testCases {
		log.Info(i, "번째 테스트 케이스")

		input, err := Decode(testCase.Input)
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Info("입력:\n", string(input))

		output, err := Decode(testCase.Output)
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Info("출력:\n", string(output))
	}

	return nil
}

func copyTestCases(submitId int, problemId int) error {
	log.Info("-----------------------")
	log.Info("테스트 케이스 복사 중...")

	if err := MakeDir("submit/" + strconv.Itoa(submitId) + "/in/"); err != nil {
		log.Fatal(err)
		return err
	}

	if err := MakeDir("submit/" + strconv.Itoa(submitId) + "/out/"); err != nil {
		log.Fatal(err)
		return err
	}

	testCaseNum := 1

	for i := 0; i < testCaseNum; i++ {
		srcInputFilePath := "problem/" + strconv.Itoa(problemId) + "/in/" + strconv.Itoa(i) + ".in"
		dstInputFilePath := "submit/" + strconv.Itoa(submitId) + "/in/" + strconv.Itoa(i) + ".in"
		srcOutputFilePath := "problem/" + strconv.Itoa(problemId) + "/out/" + strconv.Itoa(i) + ".out"
		dstOutputFilePath := "submit/" + strconv.Itoa(submitId) + "/out/" + strconv.Itoa(i) + ".out"

		if err := CopyFile(srcInputFilePath, dstInputFilePath); err != nil {
			log.Fatal(err)
			return err
		}

		if err := CopyFile(srcOutputFilePath, dstOutputFilePath); err != nil {
			log.Fatal(err)
			return err
		}
	}

	log.Info("테스트 케이스 복사 완료!")
	return nil
}

func saveTestCases(submitId int, testCases []TestCase) error {
	log.Info("-----------------------")
	log.Info("테스트 케이스 저장 중...")

	if err := MakeDir("run/" + strconv.Itoa(submitId) + "/in/"); err != nil {
		log.Fatal(err)
		return err
	}

	if err := MakeDir("run/" + strconv.Itoa(submitId) + "/out/"); err != nil {
		log.Fatal(err)
		return err
	}

	for i, testCase := range testCases {
		inputFilePath := "run/" + strconv.Itoa(submitId) + "/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := Decode(testCase.Input)
		if err != nil {
			log.Fatal(err)
			return err
		}
		if err = os.WriteFile(inputFilePath, inputContents, 0644); err != nil {
			return err
		}

		outputFilePath := "run/" + strconv.Itoa(submitId) + "/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := Decode(testCase.Output)
		if err != nil {
			log.Fatal(err)
			return err
		}
		if err = os.WriteFile(outputFilePath, outputContents, 0644); err != nil {
			return err
		}
	}

	log.Info("테스트 케이스 저장 완료!")
	return nil
}

func saveSourceCode(submitId int, code []byte, language, judgeType string) error {
	log.Info("-----------------------")
	log.Info("소스 코드 저장 중...")

	if err := MakeDir(judgeType + "/" + strconv.Itoa(submitId) + "/"); err != nil {
		log.Fatal(err)
		return err
	}

	inputFilePath := judgeType + "/" + strconv.Itoa(submitId) + "/Main." + FileExtension(language)
	if err := os.WriteFile(inputFilePath, code, 0644); err != nil {
		return err
	}

	log.Info("소스 코드 저장 완료!")
	return nil
}

func buildSource(submitId int, language, judgeType string, code []byte, buildCmd []string) error {
	if err := saveSourceCode(submitId, code, language, judgeType); err != nil {
		log.Fatal(err)
		return err
	}

	log.Info("-----------------------")
	log.Info("소스 코드 빌드 중...")
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

	log.Info("소스 코드 빌드 완료!")
	return nil
}

func judge(runCmd []string, submitId int, judgeType string) ([]bool, error) {
	testCaseNum := 1

	judgeResults := make([]bool, 0, testCaseNum)

	for i := 0; i < testCaseNum; i++ {
		log.Info("-----------------------")
		log.Info(i+1, "번째 테스트케이스 실행")

		inputFile := judgeType + "/" + strconv.Itoa(submitId) + "/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := os.ReadFile(inputFile)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		result, err := executeProgram(runCmd, inputContents)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		outputFile := judgeType + "/" + strconv.Itoa(submitId) + "/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := os.ReadFile(outputFile)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		judgeResult := checkDifference(result, outputContents)
		judgeResults = append(judgeResults, judgeResult)
	}

	return judgeResults, nil
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
	log.Info(outputContents)
	log.Info(string(outputContents))
	log.Info("실제 결과")
	log.Info(result)
	log.Info(string(result))

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
