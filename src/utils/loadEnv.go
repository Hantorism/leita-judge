package utils

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
)

func LoadEnv() error {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
