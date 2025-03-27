package repositories

import (
	"github.com/gofiber/fiber/v2/log"
	"leita/src/dataSources"
	. "leita/src/entities"
	. "leita/src/utils"
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

func (repository *ProblemRepository) GetProblemInfo(problemId int) (GetProblemInfoDTO, error) {
	db := repository.dataSource.GetDatabase()

	query := "SELECT limit_time, limit_memory FROM problem WHERE id = ?;"
	row := db.QueryRow(query, problemId)

	var dto GetProblemInfoDTO
	if err := row.Scan(&dto.TimeLimit, &dto.MemoryLimit); err != nil {
		log.Error(err)
		return GetProblemInfoDTO{}, err
	}

	return dto, nil
}

func (repository *ProblemRepository) SaveSubmitResult(dto SaveSubmitResultDTO) error {
	result := dto.Result
	usedMemory := dto.UsedMemory
	usedTime := dto.UsedTime
	submitId := dto.SubmitId

	db := repository.dataSource.GetDatabase()

	query := "UPDATE judge SET result = ?, used_memory = ?, used_time = ? WHERE id = ?;"

	if _, err := db.Exec(query, result, usedMemory, usedTime, submitId); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (repository *ProblemRepository) SaveCode(path string, code []byte) error {
	os := repository.dataSource.GetObjectStorage()
	if err := os.PutObject(path, code); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (repository *ProblemRepository) GetObjectsInFolder(folderPath string) ([]ObjectContent, error) {
	os := repository.dataSource.GetObjectStorage()
	objects, err := os.ListObjects(folderPath)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	contents := make([]ObjectContent, 0, len(objects))
	for _, object := range objects {
		content, err := os.GetObject(*object.Name)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		contents = append(contents, ObjectContent{
			Name:    *object.Name,
			Content: DecodeBase64(content),
		})
	}

	return contents, nil
}
