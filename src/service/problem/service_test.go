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

	runResults []runResult
	runCalls   int

	deleteErr   error
	deleteCalls int
}

func (f *fakeExecutor) Build(buildCmd []string) (entity.JudgeResultEnum, error) {
	return f.buildResult, f.buildErr
}

func (f *fakeExecutor) Run(runCmd []string, input []byte, timeLimit int) (entity.JudgeResultEnum, []byte, int64, int64, error) {
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
				testCaseCount: 2,
				inputs:        [][]byte{[]byte(b64("in0")), []byte(b64("in1"))},
				outputs:       [][]byte{[]byte(b64("3")), []byte(b64("7"))},
			},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCorrect,
				runResults: []runResult{
					{result: entity.JudgeCorrect, output: []byte("3"), usedTime: 100, usedMemory: 1000},
					{result: entity.JudgeCorrect, output: []byte("7"), usedTime: 300, usedMemory: 3000},
				},
			},
			wantResult:        entity.JudgeCorrect,
			wantUsedTime:      300, // 워밍업(첫 케이스 100ms) 제외 평균
			wantUsedMemory:    2000,
			wantDeleteCalls:   1,
			wantSaveCodeCalls: 1,
		},
		{
			name: "오답: 워밍업 케이스만 불일치해도 전체 결과는 WRONG",
			fileRepo: &fakeFileRepo{
				testCaseCount: 2,
				inputs:        [][]byte{[]byte(b64("in0")), []byte(b64("in1"))},
				outputs:       [][]byte{[]byte(b64("3")), []byte(b64("7"))},
			},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCorrect,
				runResults: []runResult{
					{result: entity.JudgeCorrect, output: []byte("WRONG"), usedTime: 100, usedMemory: 1000},
					{result: entity.JudgeCorrect, output: []byte("7"), usedTime: 300, usedMemory: 3000},
				},
			},
			wantResult:        entity.JudgeWrong,
			wantUsedTime:      300,
			wantUsedMemory:    2000,
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
			name: "Run 실패(타임아웃)",
			fileRepo: &fakeFileRepo{
				testCaseCount: 1,
				inputs:        [][]byte{[]byte(b64("in0"))},
				outputs:       [][]byte{[]byte(b64("3"))},
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
			service := NewService(storage, tt.fileRepo, tt.exec)

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
			},
			fileRepo: &fakeFileRepo{
				testCaseCount: 2,
				inputs:        [][]byte{[]byte(b64("in0")), []byte(b64("in1"))},
				outputs:       [][]byte{[]byte(b64("3")), []byte(b64("7"))},
			},
			exec: &fakeExecutor{
				buildResult: entity.JudgeCorrect,
				runResults: []runResult{
					{result: entity.JudgeCorrect, output: []byte("3")},
					{result: entity.JudgeCorrect, output: []byte("WRONG")},
				},
			},
			wantResults: []entity.RunProblemResult{
				{Result: entity.JudgeCorrect, Output: b64("3")},
				{Result: entity.JudgeWrong, Output: b64("WRONG")},
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
			service := NewService(&fakeStorage{}, tt.fileRepo, tt.exec)

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
