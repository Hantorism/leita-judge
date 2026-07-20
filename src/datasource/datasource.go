package datasource

import (
	"github.com/gofiber/fiber/v3/log"
	"github.com/redis/go-redis/v9"
)

type DataSource struct {
	objectStorage *ObjectStorage
	redisClient   *redis.Client
}

func NewDataSource() (*DataSource, error) {
	os, err := NewObjectStorage()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	rdb, err := NewRedisClient()
	if err != nil {
		log.Errorf("Redis initialization failed: %v", err)
		return nil, err
	}

	return &DataSource{
		objectStorage: os,
		redisClient:   rdb,
	}, nil
}

func (ds *DataSource) GetObjectStorage() *ObjectStorage {
	return ds.objectStorage
}

func (ds *DataSource) GetRedisClient() *redis.Client {
	return ds.redisClient
}
