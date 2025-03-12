package repositories

import (
	"context"

	"github.com/gofiber/fiber/v2/log"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
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

func (repository *ProblemRepository) GetObjectsInFolder(folderPath string) ([]ObjectContent, error) {
	request := objectstorage.ListObjectsRequest{
		NamespaceName: common.String(GetEnv("OS_NAMESPACE")),
		BucketName:    common.String(GetEnv("OS_BUCKET")),
		Prefix:        common.String(folderPath),
	}

	os := repository.dataSource.GetObjectStorage()
	response, err := os.Client.ListObjects(context.Background(), request)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	contents := make([]ObjectContent, 0)
	for _, object := range response.ListObjects.Objects {
		content, err := os.GetObject(*object.Name)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		contents = append(contents, ObjectContent{
			Name:    *object.Name,
			Content: content,
		})
	}

	return contents, nil
}
