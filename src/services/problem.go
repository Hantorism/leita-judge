package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/gofiber/fiber/v2/log"
	. "leita/src/entities"
	"leita/src/repositories"
	. "leita/src/utils"
)

type ProblemService struct {
	repository *repositories.ProblemRepository
}

func NewProblemService() (*ProblemService, error) {
	repository, err := repositories.NewProblemRepository()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &ProblemService{
		repository: repository,
	}, nil
}

func (service *ProblemService) SubmitProblem(dto SubmitProblemDTO) (JudgeResultEnum, error) {
	problemId := dto.ProblemId
	submitId := dto.SubmitId
	language := dto.Language
	code := dto.Code
	buildCmd := dto.BuildCmd
	runCmd := dto.RunCmd
	deleteCmd := dto.DeleteCmd

	result := JudgeUnknown

	printSubmitProblemInfo(language, submitId, problemId, code)

	if err := saveSubmitTestCases(service, submitId, problemId); err != nil {
		log.Error(err)
		return result, err
	}

	defer func() {
		saveSubmitResultDAO := SaveSubmitResultDAO{
			Result:     result.String(),
			UsedMemory: 1,
			UsedTime:   1,
			SubmitId:   submitId,
		}

		if err := service.repository.SaveSubmitResult(saveSubmitResultDAO); err != nil {
			log.Error(err)
		}
	}()

	result, err := buildSource(submitId, language, "submit", code, buildCmd)
	if err != nil {
		log.Error(err)
		return result, err
	}

	defer func(language string, deleteCmd []string) {
		if err = deleteProgram(language, deleteCmd); err != nil {
			log.Error(err)
		}
	}(language, deleteCmd)

	result, err = judgeSubmit(runCmd, submitId, "submit")
	if err != nil {
		log.Error(err)
		return result, err
	}

	return result, nil
}

func (service *ProblemService) RunProblem(dto RunProblemDTO) []RunProblemResult {
	problemId := dto.ProblemId
	submitId := dto.SubmitId
	language := dto.Language
	code := dto.Code
	testCases := dto.TestCases
	buildCmd := dto.BuildCmd
	runCmd := dto.RunCmd
	deleteCmd := dto.DeleteCmd

	result := JudgeUnknown

	if err := printRunProblemInfo(language, submitId, problemId, code, testCases); err != nil {
		log.Error(err)
		return []RunProblemResult{{Result: result, Error: err}}
	}

	if err := saveRunTestCases(submitId, testCases); err != nil {
		log.Error(err)
		return []RunProblemResult{{Result: result, Error: err}}
	}

	result, err := buildSource(submitId, language, "run", code, buildCmd)
	if err != nil {
		log.Error(err)
		return []RunProblemResult{{Result: result, Error: err}}
	}

	defer func(language string, deleteCmd []string) {
		if err = deleteProgram(language, deleteCmd); err != nil {
			log.Error(err)
		}
	}(language, deleteCmd)

	results := judgeRun(runCmd, submitId, "run")

	return results
}

func printSubmitProblemInfo(language string, submitId int, problemId int, code []byte) {
	log.Info("-----------------------")
	log.Info("언어: ", language)
	log.Info("제출 번호: ", submitId)
	log.Info("문제 번호: ", problemId)
	log.Info("코드 길이: ", len(string(code)))
	log.Info("제출 코드:\n", string(code))
}

func printRunProblemInfo(language string, submitId int, problemId int, code []byte, testCases []TestCase) error {
	log.Info("-----------------------")
	log.Info("언어: ", language)
	log.Info("제출 번호: ", submitId)
	log.Info("문제 번호: ", problemId)
	log.Info("코드 길이: ", len(string(code)))
	log.Info("제출 코드:\n", string(code))
	log.Info("테스트 케이스:")
	for i, testCase := range testCases {
		log.Info(i+1, "번째 테스트 케이스")

		input, err := Decode(testCase.Input)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Info("입력:\n", string(input))

		output, err := Decode(testCase.Output)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Info("출력:\n", string(output))
	}

	return nil
}

func saveSubmitTestCases(service *ProblemService, submitId, problemId int) error {
	log.Info("-----------------------")
	log.Info("테스트 케이스 저장 중...")

	if err := MakeDir("submit/" + strconv.Itoa(submitId) + "/in/"); err != nil {
		log.Error(err)
		return err
	}

	if err := MakeDir("submit/" + strconv.Itoa(submitId) + "/out/"); err != nil {
		log.Error(err)
		return err
	}

	testCases, err := service.repository.GetObjectsInFolder("testcases/" + strconv.Itoa(problemId))
	if err != nil {
		log.Error(err)
		return err
	}

	testCaseNum := len(testCases) / 2
	inputTestCases := testCases[:testCaseNum]
	outputTestCases := testCases[testCaseNum:]

	for i := 0; i < testCaseNum; i++ {
		inputFilePath := "submit/" + strconv.Itoa(submitId) + "/in/" + strconv.Itoa(i) + ".in"
		outputFilePath := "submit/" + strconv.Itoa(submitId) + "/out/" + strconv.Itoa(i) + ".out"

		if err = os.WriteFile(inputFilePath, inputTestCases[i].Content, 0644); err != nil {
			log.Error(err)
			return err
		}

		if err = os.WriteFile(outputFilePath, outputTestCases[i].Content, 0644); err != nil {
			log.Error(err)
			return err
		}
	}

	log.Info("테스트 케이스 저장 완료!")
	return nil
}

func saveRunTestCases(submitId int, testCases []TestCase) error {
	log.Info("-----------------------")
	log.Info("테스트 케이스 저장 중...")

	if err := MakeDir("run/" + strconv.Itoa(submitId) + "/in/"); err != nil {
		log.Error(err)
		return err
	}

	if err := MakeDir("run/" + strconv.Itoa(submitId) + "/out/"); err != nil {
		log.Error(err)
		return err
	}

	for i, testCase := range testCases {
		inputFilePath := "run/" + strconv.Itoa(submitId) + "/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := Decode(testCase.Input)
		if err != nil {
			log.Error(err)
			return err
		}
		if err = os.WriteFile(inputFilePath, inputContents, 0644); err != nil {
			log.Error(err)
			return err
		}

		outputFilePath := "run/" + strconv.Itoa(submitId) + "/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := Decode(testCase.Output)
		if err != nil {
			log.Error(err)
			return err
		}
		if err = os.WriteFile(outputFilePath, outputContents, 0644); err != nil {
			log.Error(err)
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
		log.Error(err)
		return err
	}

	inputFilePath := judgeType + "/" + strconv.Itoa(submitId) + "/Main." + FileExtension(language)
	if err := os.WriteFile(inputFilePath, code, 0644); err != nil {
		log.Error(err)
		return err
	}

	log.Info("소스 코드 저장 완료!")
	return nil
}

func buildSource(submitId int, language, judgeType string, code []byte, buildCmd []string) (JudgeResultEnum, error) {
	if err := saveSourceCode(submitId, code, language, judgeType); err != nil {
		log.Error(err)
		return JudgeUnknown, err
	}

	log.Info("-----------------------")
	log.Info("소스 코드 빌드 중...")
	if len(buildCmd) == 0 {
		log.Info(language + " 빌드 생략")
		return JudgeCorrect, nil
	}

	var stderr bytes.Buffer
	cmd := exec.Command(buildCmd[0], buildCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		compileError := fmt.Errorf("\n%w\n%s", err, stderr.String())
		log.Error(compileError)
		return JudgeCompileError, compileError
	}

	log.Info("소스 코드 빌드 완료!")
	return JudgeCorrect, nil
}

func judgeSubmit(runCmd []string, submitId int, judgeType string) (JudgeResultEnum, error) {
	testCaseNum, err := GetTestCaseNum(judgeType + "/" + strconv.Itoa(submitId) + "/in/")
	if err != nil {
		log.Error(err)
		return JudgeUnknown, err
	}

	judgeResults := make([]bool, 0, testCaseNum)

	for i := 0; i < testCaseNum; i++ {
		log.Info("-----------------------")
		log.Info(i+1, "번째 테스트케이스 실행")

		inputFile := judgeType + "/" + strconv.Itoa(submitId) + "/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := os.ReadFile(inputFile)
		if err != nil {
			log.Error(err)
			return JudgeUnknown, err
		}

		result, executeContents, err := executeProgram(runCmd, inputContents)
		if err != nil {
			log.Error(err)
			return result, err
		}

		outputFile := judgeType + "/" + strconv.Itoa(submitId) + "/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := os.ReadFile(outputFile)
		if err != nil {
			log.Error(err)
			return JudgeUnknown, err
		}

		judgeResult := checkDifference(executeContents, outputContents)
		judgeResults = append(judgeResults, judgeResult)
	}

	if !All(judgeResults) {
		log.Info("-----------------------")
		log.Info("문제를 맞추지 못했습니다.")
		return JudgeWrong, nil
	}

	log.Info("-----------------------")
	log.Info("문제를 맞췄습니다!")
	return JudgeCorrect, nil
}

func judgeRun(runCmd []string, submitId int, judgeType string) []RunProblemResult {
	testCaseNum, err := GetTestCaseNum(judgeType + "/" + strconv.Itoa(submitId) + "/in/")
	if err != nil {
		log.Error(err)
		return []RunProblemResult{{Result: JudgeUnknown, Error: err}}
	}

	results := make([]RunProblemResult, 0, testCaseNum)

	for i := 0; i < testCaseNum; i++ {
		log.Info("-----------------------")
		log.Info(i+1, "번째 테스트케이스 실행")

		inputFile := judgeType + "/" + strconv.Itoa(submitId) + "/in/" + strconv.Itoa(i) + ".in"
		inputContents, err := os.ReadFile(inputFile)
		if err != nil {
			log.Error(err)
			return []RunProblemResult{{Result: JudgeUnknown, Error: err}}
		}

		result, executeContents, err := executeProgram(runCmd, inputContents)
		if err != nil {
			log.Error(err)
			return []RunProblemResult{{Result: result, Error: err}}
		}

		outputFile := judgeType + "/" + strconv.Itoa(submitId) + "/out/" + strconv.Itoa(i) + ".out"
		outputContents, err := os.ReadFile(outputFile)
		if err != nil {
			log.Error(err)
			return []RunProblemResult{{Result: JudgeUnknown, Error: err}}
		}

		judgeResult := checkDifference(executeContents, outputContents)
		result = JudgeWrong
		if judgeResult {
			result = JudgeCorrect
		}
		log.Info(result.String())
		results = append(results, RunProblemResult{Result: result})
	}

	return results
}

func removeLineFeed(output []byte) []byte {
	if len(output) > 0 && output[len(output)-1] == 10 {
		return output[:len(output)-1]
	}

	return output
}

func executeProgram(runCmd []string, inputContents []byte) (JudgeResultEnum, []byte, error) {
	log.Info("프로그램 실행 중...")
	cmd := exec.Command(runCmd[0], runCmd[1:]...)
	cmd.Stdin = bytes.NewReader(inputContents)

	var outputBuffer bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		runtimeError := fmt.Errorf("\n%w\n%s", err, stderr.String())
		log.Error(runtimeError)
		return JudgeRuntimeError, nil, runtimeError
	}

	output := outputBuffer.Bytes()
	output = removeLineFeed(output)
	return JudgeCorrect, output, nil
}

func checkDifference(executeContents, outputContents []byte) bool {
	log.Info("예상 결과\n", outputContents, "\n", string(outputContents))
	log.Info("실제 결과\n", executeContents, "\n", string(executeContents))

	log.Info("결과를 비교 중...")
	if !bytes.Equal(executeContents, outputContents) {
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

	var stderr bytes.Buffer
	cmd := exec.Command(deleteCmd[0], deleteCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		deleteError := fmt.Errorf("\n%w\n%s", err, stderr.String())
		log.Error(deleteError)
		return deleteError
	}

	log.Info("실행 파일 삭제 완료!")
	return nil
}
