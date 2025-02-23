package main

import (
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	. "leita/src/routes"
	. "leita/src/utils"
)

// @title		Leita API Docs
// @BasePath	/api
func main() {
	initialize()

	app := fiber.New()

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(healthcheck.New())
	app.Use(swagger.New(swagger.Config{
		FilePath: "./docs/swagger.json",
		Path:     "/api/swagger",
	}))

	RegisterRoutes(app)

	log.Fatal(app.Listen(":1323"))
}

func initialize() {
	LoadEnv()

	MakeDir()
}
