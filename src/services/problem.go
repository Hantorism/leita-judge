package services

import (
	"github.com/gofiber/fiber/v2"
)

type ProblemService interface {
	JudgeProblem(c *fiber.Ctx) error
}

type problemService struct {
}

func NewProblemService() ProblemService {
	return &problemService{}
}

func (s *problemService) JudgeProblem(c *fiber.Ctx) error {
	return nil
}
