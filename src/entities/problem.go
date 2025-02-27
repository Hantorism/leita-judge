package entities

type JudgeProblemRequest struct {
	SubmitId int    `json:"submitId"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

type JudgeProblemResponse struct {
	IsSuccessful bool   `json:"isSuccessful"`
	Error        string `json:"error"`
}

type JudgeProblemDTO struct {
	ProblemId    string
	SubmitId     string
	Language     string
	Code         []byte
	Testcases    int
	RequireBuild bool
	BuildCmd     []string
	RunCmd       []string
	DeleteCmd    []string
}

type SaveJudgeResultDAO struct {
	SubmitId     int
	ProblemId    int
	Result       string
	SizeOfCode   int
	UsedLanguage string
	UsedMemory   int
	UsedTime     int
	UserId       int
}

type JudgeProblemResult struct {
	Status       int
	IsSuccessful bool
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
