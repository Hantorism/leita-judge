package dataSources

import (
	"database/sql"

	"github.com/gofiber/fiber/v2/log"
)

type DataSources struct {
	Database *sql.DB
}

func NewDataSources() *DataSources {
	db, err := NewDatabase()
	if err != nil {
		log.Fatal(err)
	}

	return &DataSources{
		Database: db,
	}
}
