package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

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

	problemInfo, err := service.repository.GetProblemInfo(problemId)
	if err != nil {
		log.Error(err)
		return result, err
	}
	timeLimit := problemInfo.TimeLimit
	memoryLimit := problemInfo.MemoryLimit
	var usedTime int64

	printSubmitProblemInfo(language, submitId, problemId, code, timeLimit, memoryLimit)

	if err = saveSubmitTestCases(service, submitId, problemId); err != nil {
		log.Error(err)
		return result, err
	}

	defer func() {
		path := filepath.Join("submits", strconv.Itoa(submitId), "Main."+FileExtension(language))
		if err = saveCode(service, path, code); err != nil {
			log.Error(err)
			return
		}
	}()

	defer func() {
		saveSubmitResultDTO := SaveSubmitResultDTO{
			Result:     result.String(),
			UsedMemory: 1,
			UsedTime: usedTime,
			SubmitId:   submitId,
		}

		log.Info("--------------------------------")
		log.Info("데이터베이스에 채점 결과 저장 중...")
		if err = service.repository.SaveSubmitResult(saveSubmitResultDTO); err != nil {
			log.Error(err)
			return
		}
		log.Info("데이터베이스에 채점 결과 저장 완료!")
	}()

	result, err = buildSource(submitId, language, "submit", code, buildCmd)
	if err != nil {
		log.Error(err)
		return result, err
	}

	defer func() {
		if err = deleteProgram(language, deleteCmd); err != nil {
			log.Error(err)
			return
		}
	}()

	result, usedTime, err = judgeSubmit(runCmd, submitId, "submit", timeLimit, memoryLimit)
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

	problemInfo, err := service.repository.GetProblemInfo(problemId)
	if err != nil {
		log.Error(err)
		return []RunProblemResult{{Result: result, Error: err}}
	}
	timeLimit := problemInfo.TimeLimit
	memoryLimit := problemInfo.MemoryLimit

	printRunProblemInfo(language, submitId, problemId, code, testCases, timeLimit, memoryLimit)

	if err = saveRunTestCases(submitId, testCases); err != nil {
		log.Error(err)
		return []RunProblemResult{{Result: result, Error: err}}
	}

	result, err = buildSource(submitId, language, "run", code, buildCmd)
	if err != nil {
		log.Error(err)
		return []RunProblemResult{{Result: result, Error: err}}
	}

	defer func() {
		if err = deleteProgram(language, deleteCmd); err != nil {
			log.Error(err)
			return
		}
	}()

	results := judgeRun(runCmd, submitId, "run", timeLimit, memoryLimit)

	return results
}

func printSubmitProblemInfo(language string, submitId, problemId int, code []byte, timeLimit, memoryLimit int) {
	log.Info("--------------------------------")
	log.Info("언어: ", language)
	log.Info("제출 번호: ", submitId)
	log.Info("문제 번호: ", problemId)
	log.Info("시간 제한: ", timeLimit, "ms")
	log.Info("메모리 제한: ", memoryLimit, "KB")
	log.Info("코드 길이: ", len(string(code)), "B")
	log.Info("제출 코드:\n", string(code))
}

func printRunProblemInfo(language string, submitId, problemId int, code []byte, testCases []TestCase, timeLimit, memoryLimit int) {
	log.Info("--------------------------------")
	log.Info("언어: ", language)
	log.Info("제출 번호: ", submitId)
	log.Info("문제 번호: ", problemId)
	log.Info("시간 제한: ", timeLimit, "ms")
	log.Info("메모리 제한: ", memoryLimit, "KB")
	log.Info("코드 길이: ", len(string(code)), "B")
	log.Info("제출 코드:\n", string(code))
	log.Info("테스트 케이스:")
	for i, testCase := range testCases {
		log.Info(i+1, "번째 테스트 케이스")

		input := DecodeBase64([]byte(testCase.Input))
		log.Info("입력:\n", string(input))

		output := DecodeBase64([]byte(testCase.Output))
		log.Info("출력:\n", string(output))
	}
}

func saveSubmitTestCases(service *ProblemService, submitId, problemId int) error {
	log.Info("--------------------------------")
	log.Info("테스트 케이스 저장 중...")

	if err := MakeDir(filepath.Join("submit", strconv.Itoa(submitId), "in")); err != nil {
		log.Error(err)
		return err
	}

	if err := MakeDir(filepath.Join("submit", strconv.Itoa(submitId), "out")); err != nil {
		log.Error(err)
		return err
	}

	//testCases, err := service.repository.GetObjectsInFolder(filepath.Join("testcases", strconv.Itoa(problemId)))
	testCases, err := service.repository.GetTestcases(problemId)
	if err != nil {
		log.Error(err)
		return err
	}

	testCaseNum := len(testCases) / 2
	inputTestCases := testCases[:testCaseNum]
	outputTestCases := testCases[testCaseNum:]

	for i := 0; i < testCaseNum; i++ {
		inputFilePath := filepath.Join("submit", strconv.Itoa(submitId), "in", strconv.Itoa(i)+".in")
		if err = os.WriteFile(inputFilePath, inputTestCases[i], 0644); err != nil {
			log.Error(err)
			return err
		}

		outputFilePath := filepath.Join("submit", strconv.Itoa(submitId), "out", strconv.Itoa(i)+".out")
		if err = os.WriteFile(outputFilePath, outputTestCases[i], 0644); err != nil {
			log.Error(err)
			return err
		}
	}

	log.Info("테스트 케이스 저장 완료!")
	return nil
}

func saveRunTestCases(submitId int, testCases []TestCase) error {
	log.Info("--------------------------------")
	log.Info("테스트 케이스 저장 중...")

	if err := MakeDir(filepath.Join("run", strconv.Itoa(submitId), "in")); err != nil {
		log.Error(err)
		return err
	}

	if err := MakeDir(filepath.Join("run", strconv.Itoa(submitId), "out")); err != nil {
		log.Error(err)
		return err
	}

	for i, testCase := range testCases {
		inputContents := DecodeBase64([]byte(testCase.Input))
		inputFilePath := filepath.Join("run", strconv.Itoa(submitId), "in", strconv.Itoa(i)+".in")
		if err := os.WriteFile(inputFilePath, inputContents, 0644); err != nil {
			log.Error(err)
			return err
		}

		outputContents := DecodeBase64([]byte(testCase.Output))
		outputFilePath := filepath.Join("run", strconv.Itoa(submitId), "out", strconv.Itoa(i)+".out")
		if err := os.WriteFile(outputFilePath, outputContents, 0644); err != nil {
			log.Error(err)
			return err
		}
	}

	log.Info("테스트 케이스 저장 완료!")
	return nil
}

func saveSourceCode(submitId int, code []byte, language, judgeType string) error {
	log.Info("--------------------------------")
	log.Info("소스 코드 저장 중...")

	if err := MakeDir(filepath.Join(judgeType, strconv.Itoa(submitId))); err != nil {
		log.Error(err)
		return err
	}

	sourceFilePath := filepath.Join(judgeType, strconv.Itoa(submitId), "Main."+FileExtension(language))
	if err := os.WriteFile(sourceFilePath, code, 0644); err != nil {
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

	log.Info("--------------------------------")
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

func judgeSubmit(runCmd []string, submitId int, judgeType string, timeLimit, memoryLimit int) (JudgeResultEnum, int64, error) {
	testCaseNum, err := GetTestCaseNum(filepath.Join(judgeType, strconv.Itoa(submitId), "in"))
	if err != nil {
		log.Error(err)
		return JudgeUnknown, 0, err
	}

	judgeResults := make([]bool, 0, testCaseNum)
	usedTimes := make([]int64, 0, testCaseNum)

	for i := 0; i < testCaseNum; i++ {
		log.Info("--------------------------------")
		log.Info(i+1, "번째 테스트케이스 실행")

		inputContents, err := os.ReadFile(filepath.Join(judgeType, strconv.Itoa(submitId), "in", strconv.Itoa(i)+".in"))
		if err != nil {
			log.Error(err)
			return JudgeUnknown, 0, err
		}

		result, executeContents, usedTime, err := executeProgram(runCmd, inputContents, timeLimit, memoryLimit)
		if err != nil {
			log.Error(err)
			return result, 0, err
		}

		outputContents, err := os.ReadFile(filepath.Join(judgeType, strconv.Itoa(submitId), "out", strconv.Itoa(i)+".out"))
		if err != nil {
			log.Error(err)
			return JudgeUnknown, 0, err
		}

		log.Info("사용 시간: ", usedTime, "ms")
		judgeResult := checkDifference(executeContents, outputContents)
		judgeResults = append(judgeResults, judgeResult)
		if i != 0 {
			usedTimes = append(usedTimes, usedTime)
		}
	}

	if !All(judgeResults) {
		log.Info("--------------------------------")
		log.Info("문제를 맞추지 못했습니다.")
		usedTime := Sum(usedTimes) / (int64(testCaseNum) - 1)
		return JudgeWrong, usedTime, nil
	}

	log.Info("--------------------------------")
	log.Info("문제를 맞췄습니다!")
	usedTime := Sum(usedTimes) / (int64(testCaseNum) - 1)
	return JudgeCorrect, usedTime, nil
}

func judgeRun(runCmd []string, submitId int, judgeType string, timeLimit, memoryLimit int) []RunProblemResult {
	testCaseNum, err := GetTestCaseNum(filepath.Join(judgeType, strconv.Itoa(submitId), "in"))
	if err != nil {
		log.Error(err)
		return []RunProblemResult{{Result: JudgeUnknown, Error: err}}
	}

	results := make([]RunProblemResult, 0, testCaseNum)

	for i := 0; i < testCaseNum; i++ {
		log.Info("--------------------------------")
		log.Info(i+1, "번째 테스트케이스 실행")

		inputContents, err := os.ReadFile(filepath.Join(judgeType, strconv.Itoa(submitId), "in", strconv.Itoa(i)+".in"))
		if err != nil {
			log.Error(err)
			return []RunProblemResult{{Result: JudgeUnknown, Error: err}}
		}

		result, executeContents, usedTime, err := executeProgram(runCmd, inputContents, timeLimit, memoryLimit)
		if err != nil {
			log.Error(err)
			return []RunProblemResult{{Result: result, Error: err}}
		}

		outputContents, err := os.ReadFile(filepath.Join(judgeType, strconv.Itoa(submitId), "out", strconv.Itoa(i)+".out"))
		if err != nil {
			log.Error(err)
			return []RunProblemResult{{Result: JudgeUnknown, Error: err}}
		}

		log.Info("사용 시간: ", usedTime, "ms")
		isSame := checkDifference(executeContents, outputContents)
		result = JudgeWrong
		if isSame {
			result = JudgeCorrect
		}
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

func executeProgram(runCmd []string, inputContents []byte, timeLimit, memoryLimit int) (JudgeResultEnum, []byte, int64, error) {
	log.Info("프로그램 실행 중...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, runCmd[0], runCmd[1:]...)
	cmd.Stdin = bytes.NewReader(inputContents)

	var outputBuffer bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		log.Error(err)
		return JudgeRuntimeError, nil, 0, err
	}

	startTime := time.Now()
	err := cmd.Wait()
	usedTime := time.Since(startTime).Milliseconds()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		log.Error(ctx.Err().Error())
		return JudgeTimeOut, nil, 0, ctx.Err()
	}

	if err != nil {
		runtimeError := fmt.Errorf("\n%w\n%s", err, stderr.String())
		log.Error(runtimeError)
		return JudgeRuntimeError, nil, 0, err
	}

	output := outputBuffer.Bytes()
	output = removeLineFeed(output)

	return JudgeCorrect, output, usedTime, nil
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
	log.Info("--------------------------------")
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

func saveCode(service *ProblemService, path string, code []byte) error {
	log.Info("--------------------------------")
	log.Info("오브젝트 스토리지에 제출 코드 저장 중...")

	if err := service.repository.SaveCode(path, EncodeBase64(code)); err != nil {
		log.Error(err)
		return err
	}

	log.Info("오브젝트 스토리지에 제출 코드 저장 완료!")
	return nil
}
