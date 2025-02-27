package repositories

import (
	"fmt"
	"strconv"

	"leita/src/dataSources"
	. "leita/src/entities"
)

type ProblemRepository interface {
	SaveJudgeResult(dto SaveJudgeResultDAO) error
}

type problemRepository struct{}

func NewProblemRepository() ProblemRepository {
	return &problemRepository{}
}

func (repository *problemRepository) SaveJudgeResult(dto SaveJudgeResultDAO) error {
	submitId := strconv.Itoa(dto.SubmitId)
	problemId := strconv.Itoa(dto.ProblemId)
	result := dto.Result
	sizeOfCode := dto.SizeOfCode
	usedLanguage := dto.UsedLanguage
	usedMemory := dto.UsedMemory
	usedTime := dto.UsedTime
	userId := strconv.Itoa(dto.UserId)

	db := dataSources.NewDataSources().Database
	defer db.Close()

	query := "INSERT INTO submits (id, problem_id, result, size_of_code, used_language, used_memory, used_time, user_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

	if _, err := db.Exec(query, submitId, problemId, result, sizeOfCode, usedLanguage, usedMemory, usedTime, userId); err != nil {
		fmt.Println(err)
	}
	return nil
}
