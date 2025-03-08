package repositories

import (
	"github.com/gofiber/fiber/v2/log"
	"leita/src/dataSources"
	. "leita/src/entities"
)

type ProblemRepository struct {
	dataSource *dataSources.DataSource
}

func NewProblemRepository() (*ProblemRepository, error) {
	dataSource, err := dataSources.NewDataSource()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &ProblemRepository{
		dataSource: dataSource,
	}, nil
}

func (repository *ProblemRepository) SaveSubmitResult(dto SaveSubmitResultDAO) error {
	result := dto.Result
	usedMemory := dto.UsedMemory
	usedTime := dto.UsedTime
	submitId := dto.SubmitId

	db := repository.dataSource.GetDatabase()

	query := "UPDATE submits SET result = ?, used_memory = ?, used_time = ? WHERE id = ?;"

	if _, err := db.Exec(query, result, usedMemory, usedTime, submitId); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
