package repositories

import (
	"github.com/gofiber/fiber/v2/log"
	"leita/src/dataSources"
	. "leita/src/entities"
)

type ProblemRepository interface {
	SaveSubmitResult(dto SaveSubmitResultDAO) error
}

type problemRepository struct {
	dataSource dataSources.DataSource
}

func NewProblemRepository() (ProblemRepository, error) {
	dataSource, err := dataSources.NewDataSource()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &problemRepository{
		dataSource: dataSource,
	}, nil
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
		log.Error(err)
	}
	return nil
}
