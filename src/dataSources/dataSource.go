package dataSources

import (
	"database/sql"

	"github.com/gofiber/fiber/v2/log"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type DataSource interface {
	GetDatabase() *sql.DB
	GetObjectStorage() *identity.IdentityClient
}

type dataSource struct {
	Database      *sql.DB
	ObjectStorage *identity.IdentityClient
}

func NewDataSource() (DataSource, error) {
	db, err := NewDatabase()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	//client, err := NewObjectStorage()
	//if err != nil {
	//	log.Fatal(err)
	//}

	return &dataSource{
		Database: db,
		//ObjectStorage: client,
	}, nil
}

func (ds *dataSource) GetDatabase() *sql.DB {
	return ds.Database
}

func (ds *dataSource) GetObjectStorage() *identity.IdentityClient {
	return ds.ObjectStorage
}
