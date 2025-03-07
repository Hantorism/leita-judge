package entities

type SubmitProblemRequest struct {
	SubmitId int    `json:"submitId"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

type SubmitProblemResponse struct {
	IsSuccessful bool   `json:"isSuccessful"`
	Result string `json:"result"`
	Error        string `json:"error"`
}

type SubmitProblemDTO struct {
	ProblemId int
	SubmitId  int
	Language  string
	Code      []byte
	BuildCmd  []string
	RunCmd    []string
	DeleteCmd []string
}

type SaveSubmitResultDAO struct {
	SubmitId   int
	Result     string
	UsedMemory int
	UsedTime   int
}

type RunProblemRequest struct {
	Language  string     `json:"language"`
	Code      string     `json:"code"`
	TestCases []TestCase `json:"testCases"`
}

type RunProblemResponse struct {
	IsSuccessful bool   `json:"isSuccessful"`
	Result       string `json:"result"`
	Error        string `json:"error"`
}

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type RunProblemDTO struct {
	ProblemId int
	SubmitId  int
	Language  string
	Code      []byte
	TestCases []TestCase
	BuildCmd  []string
	RunCmd    []string
	DeleteCmd []string
}

type SaveRunResultDAO struct {
	SubmitId   int
	Result     string
	UsedMemory int
	UsedTime   int
}

type RunProblemResult struct {
	Result JudgeResultEnum
	Error  error
}

type JudgeResultEnum int

const (
	JudgeCorrect JudgeResultEnum = iota
	JudgeWrong
	JudgeCompileError
	JudgeRuntimeError
	JudgeMemoryOut
	JudgeTimeOut
	JudgeUnknown
)

func (jr JudgeResultEnum) String() string {
	switch jr {
	case JudgeCorrect:
		return "CORRECT"
	case JudgeWrong:
		return "WRONG"
	case JudgeCompileError:
		return "COMPILE_ERROR"
	case JudgeRuntimeError:
		return "RUNTIME_ERROR"
	case JudgeMemoryOut:
		return "MEMORY_OUT"
	case JudgeTimeOut:
		return "TIME_OUT"
	default:
		return "UNKNOWN"
	}
}
