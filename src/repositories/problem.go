package repositories

import (
	"fmt"

	"leita/src/dataSources"
	. "leita/src/entities"
)

type ProblemRepository interface {
	SaveSubmitResult(dto SaveSubmitResultDAO) error
}

type problemRepository struct {
	dataSource dataSources.DataSource
}

func NewProblemRepository() ProblemRepository {
	return &problemRepository{
		dataSource: dataSources.NewDataSources(),
	}
}

func (repository *problemRepository) SaveSubmitResult(dto SaveSubmitResultDAO) error {
	result := dto.Result
	usedMemory := dto.UsedMemory
	usedTime := dto.UsedTime
	submitId := dto.SubmitId

	db := repository.dataSource.GetDatabase()
	defer db.Close()

	query := "UPDATE submits SET result = ?, used_memory = ?, used_time = ? WHERE id = ?;"

	if _, err := db.Exec(query, result, usedMemory, usedTime, submitId); err != nil {
		fmt.Println(err)
	}
	return nil
}
