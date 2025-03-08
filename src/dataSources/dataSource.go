package dataSources

import (
	"database/sql"

	"github.com/gofiber/fiber/v2/log"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type DataSource struct {
	Database      *sql.DB
	ObjectStorage *identity.IdentityClient
}

func NewDataSource() (*DataSource, error) {
	db, err := NewDatabase()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	//client, err := NewObjectStorage()
	//if err != nil {
	//	log.Error(err)
	//}

	return &DataSource{
		Database: db,
		//ObjectStorage: client,
	}, nil
}

func (ds *DataSource) GetDatabase() *sql.DB {
	return ds.Database
}

func (ds *DataSource) GetObjectStorage() *identity.IdentityClient {
	return ds.ObjectStorage
}
