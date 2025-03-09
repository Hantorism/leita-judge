package dataSources

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2/log"
	. "leita/src/utils"
)

func getDSN() (string, error) {
	dbConf := struct {
		User     string
		Password string
		Host     string
		Port     string
		Name     string
	}{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Name:     os.Getenv("DB_NAME"),
	}

	if !AllString(dbConf.User, dbConf.Password, dbConf.Host, dbConf.Port, dbConf.Name) {
		err := fmt.Errorf("invalid database configuration")
		log.Error(err)
		return "", err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		dbConf.User,
		dbConf.Password,
		dbConf.Host,
		dbConf.Port,
		dbConf.Name,
	)

	return dsn, nil
}

func NewDatabase() (*sql.DB, error) {
	dsn, err := getDSN()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		log.Error(err)
		return nil, err
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Minute * 2)

	return db, nil
}
