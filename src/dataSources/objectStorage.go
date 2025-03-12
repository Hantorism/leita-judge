package dataSources

import (
	"bytes"
	"context"
	"io"

	"github.com/gofiber/fiber/v2/log"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
	. "leita/src/utils"
)

type ObjectStorage struct {
	Client objectstorage.ObjectStorageClient
}

func NewObjectStorage() (*ObjectStorage, error) {
	config := common.DefaultConfigProvider()
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &ObjectStorage{
		Client: client,
	}, nil
}

func (os *ObjectStorage) GetObject(objectName string) ([]byte, error) {
	request := objectstorage.GetObjectRequest{
		NamespaceName: common.String(GetEnv("OS_NAMESPACE")),
		BucketName:    common.String(GetEnv("OS_BUCKET")),
		ObjectName:    common.String(objectName),
	}

	response, err := os.Client.GetObject(context.Background(), request)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer response.Content.Close()

	content, err := io.ReadAll(response.Content)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return content, nil
}

func (os *ObjectStorage) PutObject(objectName string, data []byte) error {
	request := objectstorage.PutObjectRequest{
		NamespaceName: common.String(GetEnv("OS_NAMESPACE")),
		BucketName:    common.String(GetEnv("OS_BUCKET")),
		ObjectName:    common.String(objectName),
		PutObjectBody: io.NopCloser(bytes.NewReader(data)),
	}

	_, err := os.Client.PutObject(context.Background(), request)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
