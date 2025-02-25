package models

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
