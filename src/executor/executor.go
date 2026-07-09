package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"time"

	"leita/src/cgroup"
	"leita/src/entity"

	"github.com/gofiber/fiber/v3/log"
)

type Executor interface {
	Build(buildCmd []string) (entity.JudgeResultEnum, error)
	Run(runCmd []string, input []byte, timeLimit, memoryLimit int) (entity.JudgeResultEnum, []byte, int64, int64, error)
	Delete(deleteCmd []string) error
}

type OsExecutor struct {
	monitor cgroup.Monitor
	seq     atomic.Int64
}

func NewOsExecutor(monitor cgroup.Monitor) *OsExecutor {
	if monitor == nil {
		monitor = cgroup.NewFallbackMonitor()
	}
	return &OsExecutor{monitor: monitor}
}

func (e *OsExecutor) Build(buildCmd []string) (entity.JudgeResultEnum, error) {
	log.Info("--------------------------------")
	log.Info("소스 코드 빌드 중...")

	if len(buildCmd) == 0 {
		log.Info("빌드 생략")
		return entity.JudgeCorrect, nil
	}

	var stderr bytes.Buffer
	cmd := exec.Command(buildCmd[0], buildCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		compileError := fmt.Errorf("\n%w\n%s", err, stderr.String())
		log.Error(compileError)
		return entity.JudgeCompileError, compileError
	}

	log.Info("소스 코드 빌드 완료!")
	return entity.JudgeCorrect, nil
}

func (e *OsExecutor) Run(runCmd []string, input []byte, timeLimit, memoryLimit int) (entity.JudgeResultEnum, []byte, int64, int64, error) {
	log.Info("프로그램 실행 중...")

	// 세션 이름은 파드 내 원자 카운터로 유일성을 보장한다.
	// (파드 간 유일성은 cgroup 경로의 파드 계층이 담당)
	session, err := e.monitor.NewSession(strconv.FormatInt(e.seq.Add(1), 10), memoryLimit)
	if err != nil {
		log.Warn("cgroup 세션 생성 실패, 메모리 측정 없이 실행: ", err)
		session = cgroup.NewFallbackSession()
	}
	defer func() {
		if err := session.Close(); err != nil {
			log.Warn(err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, runCmd[0], runCmd[1:]...)
	cmd.Stdin = bytes.NewReader(input)

	var outputBuffer bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &stderr

	if err := session.Apply(cmd); err != nil {
		log.Warn("cgroup 세션 적용 실패, 메모리 측정 없이 실행: ", err)
	}

	if err := cmd.Start(); err != nil {
		log.Error(err)
		return entity.JudgeRuntimeError, nil, 0, 0, err
	}

	startTime := time.Now()
	err = cmd.Wait()
	usedTime := time.Since(startTime).Milliseconds()

	usedMemory, memErr := session.PeakMemoryKB()
	if memErr != nil {
		log.Warn(memErr)
	}

	// 판정 우선순위: OOM → 타임아웃 → 런타임 에러.
	// OOM kill은 SIGKILL이라 그대로 두면 RUNTIME_ERROR로, 스래싱으로 느려진
	// 경우에는 TIME_OUT으로 오분류되므로 반드시 먼저 확인한다.
	oomKilled, oomErr := session.OOMKilled()
	if oomErr != nil {
		log.Warn(oomErr)
	}
	if oomKilled {
		memoryOutError := fmt.Errorf("메모리 제한 초과 (사용: %dKB)", usedMemory)
		log.Error(memoryOutError)
		return entity.JudgeMemoryOut, nil, usedTime, usedMemory, memoryOutError
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		log.Error(ctx.Err().Error())
		return entity.JudgeTimeOut, nil, usedTime, usedMemory, ctx.Err()
	}

	if err != nil {
		runtimeError := fmt.Errorf("\n%w\n%s", err, stderr.String())
		log.Error(runtimeError)
		return entity.JudgeRuntimeError, nil, usedTime, usedMemory, runtimeError
	}

	output := outputBuffer.Bytes()
	output = bytes.TrimRight(output, "\n\r\t ")

	return entity.JudgeCorrect, output, usedTime, usedMemory, nil
}

func (e *OsExecutor) Delete(deleteCmd []string) error {
	log.Info("--------------------------------")
	log.Info("생성된 실행 파일 삭제 중...")

	if len(deleteCmd) == 0 {
		log.Info("삭제 생략")
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
