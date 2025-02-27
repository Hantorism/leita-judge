package dataSources

import (
	"database/sql"

	"github.com/gofiber/fiber/v2/log"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type DataSources struct {
	Database      *sql.DB
	ObjectStorage *identity.IdentityClient
}

func NewDataSources() *DataSources {
	db, err := NewDatabase()
	if err != nil {
		log.Fatal(err)
	}

	client, err := NewObjectStorage()
	if err != nil {
		log.Fatal(err)
	}

	return &DataSources{
		Database:      db,
		ObjectStorage: client,
	}
}
