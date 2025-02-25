package services

import (
  "database/sql"

  . "leita/src/dataSources"
  . "leita/src/models"
)

type ProblemService interface {
  SaveJudgeResult(result JudgeResult) error
}

type problemService struct {
  db *sql.DB
}

func NewProblemService(ds *DataSources) ProblemService {
  return &problemService{db: ds.Database}
}

func (s *problemService) SaveJudgeResult(result JudgeResult) error {
  return nil
}