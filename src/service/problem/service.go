package problem

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"leita/src/entity"
	"leita/src/executor"
	"leita/src/language"
	filerepo "leita/src/repository/file"

	"github.com/gofiber/fiber/v3/log"
	"github.com/redis/go-redis/v9"
)

// storageRepository는 OCI 오브젝트 스토리지 접근을 추상화한다.
type storageRepository interface {
	SaveCode(path string, code []byte) error
	GetObjectsInFolder(path string) ([][]byte, error)
}

type Service struct {
	storage   storageRepository
	fileRepo  filerepo.Repository
	exec      executor.Executor
	redis     *redis.Client
	streamKey string
}

func NewService(
	storage storageRepository,
	fileRepo filerepo.Repository,
	exec executor.Executor,
	redisClient *redis.Client,
) *Service {
	streamKey := os.Getenv("REDIS_STREAM_KEY")
	if streamKey == "" {
		streamKey = "judge-result-stream"
	}
	return &Service{
		storage:   storage,
		fileRepo:  fileRepo,
		exec:      exec,
		redis:     redisClient,
		streamKey: streamKey,
	}
}

func (service *Service) SubmitProblem(dto entity.SubmitProblemDTO) (entity.JudgeResultEnum, int64, int64, error) {
	problemId := dto.ProblemId
	submitId := dto.SubmitId
	lang := dto.Language
	code := dto.Code
	timeLimit := dto.Limit.Time
	memoryLimit := dto.Limit.Memory

	printSubmitProblemInfo(lang, submitId, problemId, code, timeLimit, memoryLimit)

	if err := service.saveSubmitTestCases(submitId, problemId); err != nil {
		log.Error(err)
		return entity.JudgeUnknown, 0, 0, err
	}

	command := language.Commands[lang]
	buildCmd := language.ReplaceCommand(command.BuildCmd, "submit", submitId)
	runCmd := language.ReplaceCommand(command.RunCmd, "submit", submitId)
	deleteCmd := language.ReplaceCommand(command.DeleteCmd, "submit", submitId)

	if err := service.fileRepo.SaveSourceCode(submitId, code, lang, "submit"); err != nil {
		log.Error(err)
		return entity.JudgeUnknown, 0, 0, err
	}

	defer func() {
		if err := service.exec.Delete(deleteCmd); err != nil {
			log.Error(err)
		}
	}()

	defer func() {
		path := filepath.Join("submits", strconv.Itoa(submitId), language.FileName+"."+language.FileExtension(lang))
		encoded := []byte(base64.StdEncoding.EncodeToString(code))
		if err := service.storage.SaveCode(path, encoded); err != nil {
			log.Error(err)
		}
	}()

	result, err := service.exec.Build(buildCmd)
	if err != nil {
		log.Error(err)
		return result, 0, 0, err
	}

	result, usedTime, usedMemory, err := service.judgeSubmit(
		runCmd, submitId,
		language.TimeLimitWithBuffer(lang, timeLimit),
		language.MemoryLimitWithBuffer(lang, memoryLimit),
	)
	if err != nil {
		log.Error(err)
		return result, usedTime, usedMemory, err
	}

	return result, usedTime, usedMemory, nil
}

func (service *Service) RunProblem(dto entity.RunProblemDTO) []entity.RunProblemResult {
	problemId := dto.ProblemId
	lang := dto.Language
	code := dto.Code
	testCases := dto.TestCases
	timeLimit := dto.Limit.Time
	memoryLimit := dto.Limit.Memory

	minId := int(math.Pow10(11))
	maxId := int(math.Pow10(12) - 1)
	submitId := minId + rand.Intn(maxId-minId)

	printRunProblemInfo(lang, submitId, problemId, code, testCases, timeLimit, memoryLimit)

	inputs, outputs := filerepo.TestCasesFromRequest(testCases)
	if err := service.fileRepo.SaveTestCases(submitId, inputs, outputs, "run"); err != nil {
		log.Error(err)
		return []entity.RunProblemResult{{Result: entity.JudgeUnknown, Error: err}}
	}

	command := language.Commands[lang]
	buildCmd := language.ReplaceCommand(command.BuildCmd, "run", submitId)
	runCmd := language.ReplaceCommand(command.RunCmd, "run", submitId)
	deleteCmd := language.ReplaceCommand(command.DeleteCmd, "run", submitId)

	if err := service.fileRepo.SaveSourceCode(submitId, code, lang, "run"); err != nil {
		log.Error(err)
		return []entity.RunProblemResult{{Result: entity.JudgeUnknown, Error: err}}
	}

	defer func() {
		if err := service.exec.Delete(deleteCmd); err != nil {
			log.Error(err)
		}
	}()

	result, err := service.exec.Build(buildCmd)
	if err != nil {
		log.Error(err)
		return []entity.RunProblemResult{{Result: result, Error: err}}
	}

	return service.judgeRun(
		runCmd, submitId,
		language.TimeLimitWithBuffer(lang, timeLimit),
		language.MemoryLimitWithBuffer(lang, memoryLimit),
	)
}

func (service *Service) saveSubmitTestCases(submitId int, problemId string) error {
	testCasesPath := filepath.Join("problems", problemId, "testcases")
	objects, err := service.storage.GetObjectsInFolder(testCasesPath)
	if err != nil {
		log.Error(err)
		return err
	}

	testCaseNum := len(objects) / 2
	inputs := make([][]byte, testCaseNum)
	outputs := make([][]byte, testCaseNum)
	for i := 0; i < testCaseNum; i++ {
		inputs[i] = objects[i*2]
		outputs[i] = objects[i*2+1]
	}

	return service.fileRepo.SaveTestCases(submitId, inputs, outputs, "submit")
}

func (service *Service) judgeSubmit(runCmd []string, submitId int, timeLimit, memoryLimit int) (entity.JudgeResultEnum, int64, int64, error) {
	testCaseNum, err := service.fileRepo.GetTestCaseCount(submitId, "submit")
	if err != nil {
		log.Error(err)
		return entity.JudgeUnknown, 0, 0, err
	}
	if testCaseNum < 1 {
		return entity.JudgeUnknown, 0, 0, errors.New("not enough testcases")
	}

	judgeResults := make([]bool, 0, testCaseNum)
	usedTimes := make([]int64, 0, testCaseNum)
	usedMemories := make([]int64, 0, testCaseNum)

	for i := 0; i < testCaseNum; i++ {
		log.Info("--------------------------------")
		log.Info(i+1, "번째 테스트케이스 실행")

		inputData, err := service.fileRepo.ReadInput(submitId, i, "submit")
		if err != nil {
			log.Error(err)
			return entity.JudgeUnknown, 0, 0, err
		}
		inputContents, err := base64.StdEncoding.DecodeString(string(inputData))
		if err != nil {
			log.Error(err)
			return entity.JudgeUnknown, 0, 0, err
		}

		result, executeContents, usedTime, usedMemory, err := service.exec.Run(runCmd, inputContents, timeLimit, memoryLimit)
		if err != nil {
			log.Error(err)
			return result, usedTime, usedMemory, err
		}

		outputData, err := service.fileRepo.ReadOutput(submitId, i, "submit")
		if err != nil {
			log.Error(err)
			return entity.JudgeUnknown, 0, 0, err
		}
		outputContents, err := base64.StdEncoding.DecodeString(string(outputData))
		if err != nil {
			log.Error(err)
			return entity.JudgeUnknown, 0, 0, err
		}

		log.Info("사용 시간: ", usedTime, "ms")
		log.Info("사용 메모리: ", usedMemory, "KB")
		judgeResults = append(judgeResults, checkDifference(executeContents, outputContents))
		usedTimes = append(usedTimes, usedTime)
		usedMemories = append(usedMemories, usedMemory)
	}

	usedTime := averageExcludingWarmup(usedTimes)
	usedMemory := averageExcludingWarmup(usedMemories)

	if !allTrue(judgeResults) {
		printJudgeSubmitResult(false, usedTime, usedMemory)
		return entity.JudgeWrong, usedTime, usedMemory, nil
	}

	printJudgeSubmitResult(true, usedTime, usedMemory)
	return entity.JudgeCorrect, usedTime, usedMemory, nil
}

func (service *Service) judgeRun(runCmd []string, submitId int, timeLimit, memoryLimit int) []entity.RunProblemResult {
	testCaseNum, err := service.fileRepo.GetTestCaseCount(submitId, "run")
	if err != nil {
		log.Error(err)
		return []entity.RunProblemResult{{Result: entity.JudgeUnknown, Error: err}}
	}

	results := make([]entity.RunProblemResult, 0, testCaseNum)

	for i := 0; i < testCaseNum; i++ {
		log.Info("--------------------------------")
		log.Info(i+1, "번째 테스트케이스 실행")

		inputData, err := service.fileRepo.ReadInput(submitId, i, "run")
		if err != nil {
			log.Error(err)
			return []entity.RunProblemResult{{Result: entity.JudgeUnknown, Error: err}}
		}
		inputContents, err := base64.StdEncoding.DecodeString(string(inputData))
		if err != nil {
			log.Error(err)
			return []entity.RunProblemResult{{Result: entity.JudgeUnknown, Error: err}}
		}

		result, executeContents, usedTime, usedMemory, err := service.exec.Run(runCmd, inputContents, timeLimit, memoryLimit)
		if err != nil {
			log.Error(err)
			return []entity.RunProblemResult{{Result: result, Error: err}}
		}

		outputData, err := service.fileRepo.ReadOutput(submitId, i, "run")
		if err != nil {
			log.Error(err)
			return []entity.RunProblemResult{{Result: entity.JudgeUnknown, Error: err}}
		}
		outputContents, err := base64.StdEncoding.DecodeString(string(outputData))
		if err != nil {
			log.Error(err)
			return []entity.RunProblemResult{{Result: entity.JudgeUnknown, Error: err}}
		}

		log.Info("사용 시간: ", usedTime, "ms")
		log.Info("사용 메모리: ", usedMemory, "KB")

		judgeResult := entity.JudgeWrong
		if checkDifference(executeContents, outputContents) {
			judgeResult = entity.JudgeCorrect
		}
		results = append(results, entity.RunProblemResult{
			Result: judgeResult,
			Output: base64.StdEncoding.EncodeToString(executeContents),
		})
	}

	return results
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

func allTrue(s []bool) bool {
	for _, v := range s {
		if !v {
			return false
		}
	}
	return true
}

func sumInt64(s []int64) int64 {
	var sum int64
	for _, v := range s {
		sum += v
	}
	return sum
}

// averageExcludingWarmup은 첫 번째 테스트케이스(워밍업)를 제외한 평균을 구한다.
// 시간은 JIT/캐시 예열, 메모리는 페이지 캐시 첫 적재 비용이 첫 케이스에 몰리기 때문이다.
func averageExcludingWarmup(s []int64) int64 {
	if len(s) > 1 {
		return sumInt64(s[1:]) / int64(len(s)-1)
	}
	if len(s) == 1 {
		return s[0]
	}
	return 0
}

func printSubmitProblemInfo(lang string, submitId int, problemId string, code []byte, timeLimit, memoryLimit int) {
	log.Info("--------------------------------")
	log.Info("언어: ", lang)
	log.Info("제출 번호: ", submitId)
	log.Info("문제 번호: ", problemId)
	log.Info("시간 제한: ", timeLimit, "ms")
	log.Info("메모리 제한: ", memoryLimit, "KB")
	log.Info("코드 길이: ", len(string(code)), "B")
	log.Info("제출 코드:\n", string(code))
}

func printRunProblemInfo(lang string, submitId int, problemId string, code []byte, testCases []entity.TestCase, timeLimit, memoryLimit int) {
	log.Info("--------------------------------")
	log.Info("언어: ", lang)
	log.Info("제출 번호: ", submitId)
	log.Info("문제 번호: ", problemId)
	log.Info("시간 제한: ", timeLimit, "ms")
	log.Info("메모리 제한: ", memoryLimit, "KB")
	log.Info("코드 길이: ", len(string(code)), "B")
	log.Info("제출 코드:\n", string(code))
	log.Info("테스트 케이스:")
	for i, tc := range testCases {
		log.Info(i+1, "번째 테스트 케이스")
		log.Info("입력:\n", tc.Input)
		log.Info("출력:\n", tc.Output)
	}
}

func printJudgeSubmitResult(isCorrect bool, usedTime, usedMemory int64) {
	log.Info("--------------------------------")
	if isCorrect {
		log.Info("문제를 맞췄습니다!")
	} else {
		log.Info("문제를 맞추지 못했습니다.")
	}
	log.Info("평균 사용 시간: ", usedTime, "ms")
	log.Info("평균 사용 메모리: ", usedMemory, "KB")
}

func (service *Service) PublishJudgeResult(submitId int, result entity.JudgeResultEnum, usedTime, usedMemory int64, errStr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	values := map[string]interface{}{
		"submitId":   strconv.Itoa(submitId),
		"result":     result.String(),
		"error":      errStr,
		"usedMemory": strconv.FormatInt(usedMemory, 10),
		"usedTime":   strconv.FormatInt(usedTime, 10),
	}

	err := service.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: service.streamKey,
		Values: values,
	}).Err()

	if err != nil {
		log.Errorf("XAdd error: %v", err)
		return err
	}

	log.Infof("Successfully published judge result for submit %d: %s", submitId, result.String())
	return nil
}
