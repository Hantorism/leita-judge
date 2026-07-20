package entity

type SubmitProblemRequest struct {
	SubmitId int    `json:"submitId"`
	Language string `json:"language"`
	Code     string `json:"code"`
	Limit    Limit  `json:"limit"`
}

type SubmitProblemResponse struct {
	Result     string `json:"result"`
	Error      string `json:"error"`
	UsedTime   int64  `json:"usedTime"`
	UsedMemory int64  `json:"usedMemory"`
}

type SubmitProblemReceiptResponse struct {
	Status string `json:"status"`
}

type SubmitProblemDTO struct {
	ProblemId string
	SubmitId  int
	Language  string
	Code      []byte
	Limit     Limit
}

type SaveSubmitResultDTO struct {
	SubmitId   int
	Result     string
	UsedMemory int64
	UsedTime   int64
}

type RunProblemRequest struct {
	Language  string     `json:"language"`
	Code      string     `json:"code"`
	Limit     Limit      `json:"limit"`
	TestCases []TestCase `json:"testCases"`
}

type RunProblemResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
	Output string `json:"output"`
}

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type RunProblemDTO struct {
	ProblemId string
	Language  string
	Code      []byte
	Limit     Limit
	TestCases []TestCase
}

type RunProblemResult struct {
	Result JudgeResultEnum
	Error  error
	Output string
}

type JudgeResultEnum int

const (
	JudgeUnknown JudgeResultEnum = iota
	JudgeCorrect
	JudgeWrong
	JudgeCompileError
	JudgeRuntimeError
	JudgeMemoryOut
	JudgeTimeOut
)

var judgeResultStrings = []string{
	"UNKNOWN",
	"CORRECT",
	"WRONG",
	"COMPILE_ERROR",
	"RUNTIME_ERROR",
	"MEMORY_OUT",
	"TIME_OUT",
}

func (jr JudgeResultEnum) String() string {
	if jr < 0 || int(jr) >= len(judgeResultStrings) {
		return "UNKNOWN"
	}
	return judgeResultStrings[jr]
}

type Limit struct {
	Time   int `json:"time"`
	Memory int `json:"memory"`
}
