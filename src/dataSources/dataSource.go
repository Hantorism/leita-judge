package dataSources

import (
	"database/sql"

	"github.com/gofiber/fiber/v2/log"
)

type DataSource struct {
	database      *sql.DB
	objectStorage *ObjectStorage
}

func NewDataSource() (*DataSource, error) {
	db, err := NewDatabase()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	os, err := NewObjectStorage()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &DataSource{
		database:      db,
		objectStorage: os,
	}, nil
}

func (ds *DataSource) GetDatabase() *sql.DB {
	return ds.database
}

func (ds *DataSource) GetObjectStorage() *ObjectStorage {
	return ds.objectStorage
}
