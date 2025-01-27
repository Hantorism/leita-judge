package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	. "leita/src/routes"
	. "leita/src/utils"
)

func main() {
	initialize()

	app := fiber.New()

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(healthcheck.NewHealthChecker())

	RegisterRoutes(app)

	log.Fatal(app.Listen(":1323"))
}

func initialize() {
	MakeDir()
}
