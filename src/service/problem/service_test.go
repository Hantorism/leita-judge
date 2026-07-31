package problem

import (
	"encoding/base64"
	"errors"
	"testing"

	"leita/src/entity"
)

func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

type runResult struct {
	result     entity.JudgeResultEnum
	output     []byte
	usedTime   int64
	usedMemory int64
	err        error
}

type fakeExecutor struct {
	buildResult entity.JudgeResultEnum
	buildErr    error

	runResults     []runResult
	runCalls       int
	gotTimeLimit   int
	gotMemoryLimit int

	deleteErr   error
	deleteCalls int
}

func (f *fakeExecutor) Build(buildCmd []string) (entity.JudgeResultEnum, error) {
	return f.buildResult, f.buildErr
}

func (f *fakeExecutor) Run(runCmd []string, input []byte, timeLimit, memoryLimit int) (entity.JudgeResultEnum, []byte, int64, int64, error) {
	f.gotTimeLimit = timeLimit
	f.gotMemoryLimit = memoryLimit
	r := f.runResults[f.runCalls]
	f.runCalls++
	return r.result, r.output, r.usedTime, r.usedMemory, r.err
}

func (f *fakeExecutor) Delete(deleteCmd []string) error {
	f.deleteCalls++
	return f.deleteErr
}

type fakeFileRepo struct {
	saveSourceCodeErr error
	saveTestCasesErr  error

	testCaseCount    int
	testCaseCountErr error

	inputs  [][]byte
	outputs [][]byte
}

func (f *fakeFileRepo) SaveSourceCode(submitId int, code []byte, lang, judgeType string) error {
	return f.saveSourceCodeErr
}

func (f *fakeFileRepo) SaveTestCases(submitId int, inputs, outputs [][]byte, judgeType string) error {
	return f.saveTestCasesErr
}

func (f *fakeFileRepo) ReadInput(submitId int, index int, judgeType string) ([]byte, error) {
	return f.inputs[index], nil
}

func (f *fakeFileRepo) ReadOutput(submitId int, index int, judgeType string) ([]byte, error) {
	return f.outputs[index], nil
}

func (f *fakeFileRepo) GetTestCaseCount(submitId int, judgeType string) (int, error) {
	return f.testCaseCount, f.testCaseCountErr
}

type fakeStorage struct {
	objectsErr error

	saveCodeErr   error
	saveCodeCalls int
}

func (f *fakeStorage) SaveCode(path string, code []byte) error {
	f.saveCodeCalls++
	return f.saveCodeErr
}

func (f *fakeStorage) GetObjectsInFolder(path string) ([][]byte, error) {
	return nil, f.objectsErr
}

func TestSubmitProblem(t *testing.T) {
	tests := []struct {
		name string

		fileRepo *fakeFileRepo
		exec     *fakeExecutor

		wantResult        entity.JudgeResultEnum
		wantErr           bool
		wantUsedTime      int64
		wantUsedMemory    int64
		wantDeleteCalls   int
		wantSaveCodeCalls int
	}{
		{
			name: "정답: 모든 테스트케이스 일치",
			fileRepo: &fakeFileRepo{
				testCaseCount: 5,
				inputs:        [][]byte{[]byte(b64("in0")), []byte(b64("in1")), []byte(b64("in2")), []byte(b64("in3")), []byte(b64("in4"))},
				outputs:       [][]byte{[]byte(b64("3")), []byte(b64("7")), []byte(b64("11")), []byte(b64("15")), []byte(b64("19"))},
			},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCorrect,
				runResults: []runResult{
					{result: entity.JudgeCorrect, output: []byte("3"), usedTime: 100, usedMemory: 1000},
					{result: entity.JudgeCorrect, output: []byte("7"), usedTime: 200, usedMemory: 2000},
					{result: entity.JudgeCorrect, output: []byte("11"), usedTime: 300, usedMemory: 3000},
					{result: entity.JudgeCorrect, output: []byte("15"), usedTime: 400, usedMemory: 4000},
					{result: entity.JudgeCorrect, output: []byte("19"), usedTime: 500, usedMemory: 5000},
				},
			},
			wantResult:        entity.JudgeCorrect,
			wantUsedTime:      350,  // 워밍업(첫 케이스 100ms) 제외 평균 (200+300+400+500)/4
			wantUsedMemory:    3500, // 메모리도 워밍업(첫 케이스 1000KB) 제외 평균
			wantDeleteCalls:   1,
			wantSaveCodeCalls: 1,
		},
		{
			name: "오답: 워밍업 케이스만 불일치해도 전체 결과는 WRONG",
			fileRepo: &fakeFileRepo{
				testCaseCount: 5,
				inputs:        [][]byte{[]byte(b64("in0")), []byte(b64("in1")), []byte(b64("in2")), []byte(b64("in3")), []byte(b64("in4"))},
				outputs:       [][]byte{[]byte(b64("3")), []byte(b64("7")), []byte(b64("11")), []byte(b64("15")), []byte(b64("19"))},
			},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCorrect,
				runResults: []runResult{
					{result: entity.JudgeCorrect, output: []byte("WRONG"), usedTime: 100, usedMemory: 1000},
					{result: entity.JudgeCorrect, output: []byte("7"), usedTime: 200, usedMemory: 2000},
					{result: entity.JudgeCorrect, output: []byte("11"), usedTime: 300, usedMemory: 3000},
					{result: entity.JudgeCorrect, output: []byte("15"), usedTime: 400, usedMemory: 4000},
					{result: entity.JudgeCorrect, output: []byte("19"), usedTime: 500, usedMemory: 5000},
				},
			},
			wantResult:        entity.JudgeWrong,
			wantUsedTime:      350,
			wantUsedMemory:    3500,
			wantDeleteCalls:   1,
			wantSaveCodeCalls: 1,
		},
		{
			name:     "Build 실패: Delete/SaveCode는 그래도 호출된다",
			fileRepo: &fakeFileRepo{},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCompileError,
				buildErr:    errors.New("compile error"),
			},
			wantResult:        entity.JudgeCompileError,
			wantErr:           true,
			wantDeleteCalls:   1,
			wantSaveCodeCalls: 1,
		},
		{
			name: "Run 실패(타임아웃): 첫 케이스에서 중단",
			fileRepo: &fakeFileRepo{
				testCaseCount: 5,
				inputs:        [][]byte{[]byte(b64("in0")), []byte(b64("in1")), []byte(b64("in2")), []byte(b64("in3")), []byte(b64("in4"))},
				outputs:       [][]byte{[]byte(b64("3")), []byte(b64("7")), []byte(b64("11")), []byte(b64("15")), []byte(b64("19"))},
			},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCorrect,
				runResults: []runResult{
					{result: entity.JudgeTimeOut, err: errors.New("timeout")},
				},
			},
			wantResult:        entity.JudgeTimeOut,
			wantErr:           true,
			wantDeleteCalls:   1,
			wantSaveCodeCalls: 1,
		},
		{
			name:     "테스트케이스 0개",
			fileRepo: &fakeFileRepo{testCaseCount: 0},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCorrect,
			},
			wantResult:        entity.JudgeUnknown,
			wantErr:           true,
			wantDeleteCalls:   1,
			wantSaveCodeCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &fakeStorage{}
			// 채점 결과 발행(PublishJudgeResult)은 핸들러에서만 호출되므로
			// 채점 로직 테스트에서는 redis 클라이언트가 필요 없다.
			service := NewService(storage, tt.fileRepo, tt.exec, nil)

			result, usedTime, usedMemory, err := service.SubmitProblem(entity.SubmitProblemDTO{
				ProblemId: "1",
				SubmitId:  1,
				Language:  "PYTHON",
				Code:      []byte("print(42)"),
				Limit:     entity.Limit{Time: 1000, Memory: 65536},
			})

			if result != tt.wantResult {
				t.Errorf("result = %v, want %v", result, tt.wantResult)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if usedTime != tt.wantUsedTime {
				t.Errorf("usedTime = %d, want %d", usedTime, tt.wantUsedTime)
			}
			if usedMemory != tt.wantUsedMemory {
				t.Errorf("usedMemory = %d, want %d", usedMemory, tt.wantUsedMemory)
			}
			if tt.exec.deleteCalls != tt.wantDeleteCalls {
				t.Errorf("Delete calls = %d, want %d", tt.exec.deleteCalls, tt.wantDeleteCalls)
			}
			if storage.saveCodeCalls != tt.wantSaveCodeCalls {
				t.Errorf("SaveCode calls = %d, want %d", storage.saveCodeCalls, tt.wantSaveCodeCalls)
			}
			// PYTHON 버퍼(시간 ×3+2초, 메모리 +32MB)가 executor까지 전달돼야 한다.
			if tt.exec.runCalls > 0 {
				if tt.exec.gotTimeLimit != 1000*3+2000 {
					t.Errorf("timeLimit = %d, want %d (제한 × PYTHON 배수 + 가산)", tt.exec.gotTimeLimit, 1000*3+2000)
				}
				if tt.exec.gotMemoryLimit != 65536+32*1024 {
					t.Errorf("memoryLimit = %d, want %d (제한 + PYTHON 버퍼)", tt.exec.gotMemoryLimit, 65536+32*1024)
				}
			}
		})
	}
}

func TestRunProblem(t *testing.T) {
	tests := []struct {
		name string

		testCases []entity.TestCase
		fileRepo  *fakeFileRepo
		exec      *fakeExecutor

		wantResults     []entity.RunProblemResult
		wantErr         bool
		wantDeleteCalls int
	}{
		{
			name: "정답/오답 혼합",
			testCases: []entity.TestCase{
				{Input: b64("in0"), Output: b64("3")},
				{Input: b64("in1"), Output: b64("7")},
				{Input: b64("in2"), Output: b64("11")},
				{Input: b64("in3"), Output: b64("15")},
				{Input: b64("in4"), Output: b64("19")},
			},
			fileRepo: &fakeFileRepo{
				testCaseCount: 5,
				inputs:        [][]byte{[]byte(b64("in0")), []byte(b64("in1")), []byte(b64("in2")), []byte(b64("in3")), []byte(b64("in4"))},
				outputs:       [][]byte{[]byte(b64("3")), []byte(b64("7")), []byte(b64("11")), []byte(b64("15")), []byte(b64("19"))},
			},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCorrect,
				runResults: []runResult{
					{result: entity.JudgeCorrect, output: []byte("3")},
					{result: entity.JudgeCorrect, output: []byte("WRONG")},
					{result: entity.JudgeCorrect, output: []byte("11")},
					{result: entity.JudgeCorrect, output: []byte("15")},
					{result: entity.JudgeCorrect, output: []byte("WRONG2")},
				},
			},
			wantResults: []entity.RunProblemResult{
				{Result: entity.JudgeCorrect, Output: b64("3")},
				{Result: entity.JudgeWrong, Output: b64("WRONG")},
				{Result: entity.JudgeCorrect, Output: b64("11")},
				{Result: entity.JudgeCorrect, Output: b64("15")},
				{Result: entity.JudgeWrong, Output: b64("WRONG2")},
			},
			wantDeleteCalls: 1,
		},
		{
			name:      "Build 실패: Delete는 그래도 호출된다",
			testCases: []entity.TestCase{{Input: b64("in0"), Output: b64("3")}},
			fileRepo:  &fakeFileRepo{},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCompileError,
				buildErr:    errors.New("compile error"),
			},
			wantResults: []entity.RunProblemResult{
				{Result: entity.JudgeCompileError},
			},
			wantErr:         true,
			wantDeleteCalls: 1,
		},
		{
			name:      "테스트케이스 저장 실패: Delete는 호출되지 않는다",
			testCases: []entity.TestCase{{Input: b64("in0"), Output: b64("3")}},
			fileRepo: &fakeFileRepo{
				saveTestCasesErr: errors.New("disk full"),
			},
			exec: &fakeExecutor{},
			wantResults: []entity.RunProblemResult{
				{Result: entity.JudgeUnknown},
			},
			wantErr:         true,
			wantDeleteCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(&fakeStorage{}, tt.fileRepo, tt.exec, nil)

			results := service.RunProblem(entity.RunProblemDTO{
				ProblemId: "1",
				Language:  "PYTHON",
				Code:      []byte("print(42)"),
				Limit:     entity.Limit{Time: 1000, Memory: 65536},
				TestCases: tt.testCases,
			})

			if len(results) != len(tt.wantResults) {
				t.Fatalf("len(results) = %d, want %d", len(results), len(tt.wantResults))
			}
			for i, want := range tt.wantResults {
				got := results[i]
				if got.Result != want.Result {
					t.Errorf("results[%d].Result = %v, want %v", i, got.Result, want.Result)
				}
				if got.Output != want.Output {
					t.Errorf("results[%d].Output = %q, want %q", i, got.Output, want.Output)
				}
				if (got.Error != nil) != tt.wantErr {
					t.Errorf("results[%d].Error = %v, wantErr %v", i, got.Error, tt.wantErr)
				}
			}
			if tt.exec.deleteCalls != tt.wantDeleteCalls {
				t.Errorf("Delete calls = %d, want %d", tt.exec.deleteCalls, tt.wantDeleteCalls)
			}
		})
	}
}

// 미지원 언어는 파일 저장·빌드·실행 이전에 걸러져야 한다.
// 그냥 진행하면 빈 실행 커맨드로 executor에서 인덱스 패닉이 난다.
func TestUnsupportedLanguage(t *testing.T) {
	limit := entity.Limit{Time: 1000, Memory: 65536}

	t.Run("SubmitProblem은 부작용 없이 에러를 반환한다", func(t *testing.T) {
		exec := &fakeExecutor{}
		storage := &fakeStorage{}
		service := NewService(storage, &fakeFileRepo{}, exec, nil)

		result, _, _, err := service.SubmitProblem(entity.SubmitProblemDTO{
			ProblemId: "1",
			SubmitId:  1,
			Language:  "HASKELL",
			Code:      []byte("main = print 42"),
			Limit:     limit,
		})

		if result != entity.JudgeUnknown {
			t.Errorf("result = %v, want UNKNOWN", result)
		}
		if err == nil {
			t.Error("err = nil, 미지원 언어는 에러여야 한다")
		}
		if exec.runCalls != 0 || exec.deleteCalls != 0 {
			t.Errorf("실행(%d)/삭제(%d)가 시도됨 — 언어 확인 전에 진행하면 안 된다", exec.runCalls, exec.deleteCalls)
		}
		if storage.saveCodeCalls != 0 {
			t.Errorf("SaveCode가 %d회 호출됨 — 미지원 언어는 저장도 하면 안 된다", storage.saveCodeCalls)
		}
	})

	t.Run("RunProblem도 동일하게 걸러진다", func(t *testing.T) {
		exec := &fakeExecutor{}
		service := NewService(&fakeStorage{}, &fakeFileRepo{}, exec, nil)

		results := service.RunProblem(entity.RunProblemDTO{
			ProblemId: "1",
			Language:  "HASKELL",
			Code:      []byte("main = print 42"),
			Limit:     limit,
			TestCases: []entity.TestCase{{Input: b64("1 2"), Output: b64("3")}},
		})

		if len(results) != 1 {
			t.Fatalf("len(results) = %d, want 1", len(results))
		}
		if results[0].Result != entity.JudgeUnknown {
			t.Errorf("result = %v, want UNKNOWN", results[0].Result)
		}
		if results[0].Error == nil {
			t.Error("Error = nil, 미지원 언어는 에러여야 한다")
		}
		if exec.runCalls != 0 || exec.deleteCalls != 0 {
			t.Errorf("실행(%d)/삭제(%d)가 시도됨", exec.runCalls, exec.deleteCalls)
		}
	})
}
