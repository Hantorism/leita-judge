package entities

type SubmitProblemRequest struct {
	SubmitId int    `json:"submitId"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

type SubmitProblemResponse struct {
	IsSuccessful bool   `json:"isSuccessful"`
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

type SubmitProblemResult struct {
	Status       int
	IsSuccessful bool
	Error        error
}

type RunProblemRequest struct {
	Language  string     `json:"language"`
	Code      string     `json:"code"`
	TestCases []TestCase `json:"testCases"`
}

type RunProblemResponse struct {
	IsSuccessful []bool `json:"isSuccessful"`
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
	Status       int
	IsSuccessful []bool
	Error        error
}

type JudgeResultEnum int

const (
	JudgePass JudgeResultEnum = iota
	JudgeFail
	JudgeError
)

func (jr JudgeResultEnum) String() string {
	switch jr {
	case JudgeError:
		return "채점 중 이상이 있습니다."
	case JudgeFail:
		return "문제를 틀렸습니다."
	case JudgePass:
		return "문제를 맞췄습니다!"
	default:
		return "error"
	}
}
