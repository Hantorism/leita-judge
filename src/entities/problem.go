package entities

type SubmitProblemRequest struct {
	SubmitId int    `json:"submitId"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

type SubmitProblemResponse struct {
	Result       string `json:"result"`
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

type RunProblemResult struct {
	Result JudgeResultEnum
	Error  error
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

func (jr JudgeResultEnum) String() string {
	return map[JudgeResultEnum]string{
		JudgeUnknown:      "UNKNOWN",
		JudgeCorrect:      "CORRECT",
		JudgeWrong:        "WRONG",
		JudgeCompileError: "COMPILE_ERROR",
		JudgeRuntimeError: "RUNTIME_ERROR",
		JudgeMemoryOut:    "MEMORY_OUT",
		JudgeTimeOut:      "TIME_OUT",
	}[jr]
}

type ObjectContent struct {
	Name    string
	Content []byte
}
