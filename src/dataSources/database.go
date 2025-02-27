package dataSources

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	. "leita/src/functions"
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
		return "", fmt.Errorf("Invalid Database configuration")
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
		return nil, err
	}

	var db *sql.DB
	if db, err = sql.Open("mysql", dsn); err != nil {
		return nil, fmt.Errorf("Database 연결 실패: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("Database 핑 테스트 실패: %w", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Minute * 2)

	return db, nil
}
