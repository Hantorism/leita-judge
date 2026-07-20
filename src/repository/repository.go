package repository

import (
	"leita/src/datasource"
	filerepo "leita/src/repository/file"
	"leita/src/repository/problem"

	"github.com/gofiber/fiber/v3/log"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	ProblemRepository *problem.Repository
	FileRepository    filerepo.Repository
	RedisClient       *redis.Client
}

func NewRepository() (*Repository, error) {
	dataSource, err := datasource.NewDataSource()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &Repository{
		ProblemRepository: problem.NewRepository(dataSource),
		FileRepository:    filerepo.NewLocalRepository(),
		RedisClient:       dataSource.GetRedisClient(),
	}, nil
}
