package problem

import (
	"leita/src/entity"
)

// Service는 핸들러가 의존하는 서비스 인터페이스다.
type Service interface {
	SubmitProblem(dto entity.SubmitProblemDTO) (entity.JudgeResultEnum, int64, int64, error)
	RunProblem(dto entity.RunProblemDTO) []entity.RunProblemResult
	PublishJudgeResult(submitId int, result entity.JudgeResultEnum, usedTime, usedMemory int64, errStr string) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}
