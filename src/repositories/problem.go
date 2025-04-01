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

func (repository *ProblemRepository) GetProblemInfo(problemId int) (GetProblemInfoDAO, error) {
	db := repository.dataSource.GetDatabase()

	query := "SELECT limit_time, limit_memory FROM problem WHERE id = ?;"
	row := db.QueryRow(query, problemId)

	var dto GetProblemInfoDAO
	if err := row.Scan(&dto.TimeLimit, &dto.MemoryLimit); err != nil {
		log.Error(err)
		return GetProblemInfoDAO{}, err
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

func (repository *ProblemRepository) GetObjectsInFolder(folderPath string) ([][]byte, error) {
	os := repository.dataSource.GetObjectStorage()
	objects, err := os.ListObjects(folderPath)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	contents := make([][]byte, 0, len(objects))
	for _, object := range objects {
		content, err := os.GetObject(*object.Name)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		contents = append(contents, DecodeBase64(content))
	}

	return contents, nil
}

// 테스트케이스 임시로 db에서 가져오기
func (repository *ProblemRepository) GetTestcases(problemId int) ([][]byte, error) {
	db := repository.dataSource.GetDatabase()

	query := "SELECT input, output FROM problem_test_cases WHERE problem_id = ?;"
	rows, err := db.Query(query, problemId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()

	testCases := make([][]byte, 0)
	inputs := make([][]byte, 0)
	outputs := make([][]byte, 0)
	for rows.Next() {
		var input, output []byte
		if err = rows.Scan(&input, &output); err != nil {
			log.Error(err)
			return nil, err
		}
		inputs = append(inputs, DecodeBase64(input))
		outputs = append(outputs, DecodeBase64(output))
	}
	if err = rows.Err(); err != nil {
		log.Error(err)
		return nil, err
	}
	testCases = append(testCases, inputs...)
	testCases = append(testCases, outputs...)

	return testCases, nil
}
