package service

import (
	"leita/src/cgroup"
	"leita/src/executor"
	"leita/src/repository"
	"leita/src/service/problem"

	"github.com/gofiber/fiber/v3/log"
)

type Service struct {
	ProblemService *problem.Service
}

func NewService() (*Service, error) {
	repository, err := repository.NewRepository()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	monitor, err := cgroup.Setup()
	if err != nil {
		log.Warn("cgroup 초기화 실패, 메모리 측정 없이 동작합니다: ", err)
	}

	exec := executor.NewOsExecutor(monitor)
	problemService := problem.NewService(repository.ProblemRepository, repository.FileRepository, exec)

	return &Service{
		ProblemService: problemService,
	}, nil
}
