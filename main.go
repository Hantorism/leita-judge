package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	. "leita/src/routes"
	. "leita/src/utils"
)

func main() {
	initialize()

	//app := fiber.New(fiber.Config{
	//	ErrorHandler: func(ctx fiber.Ctx, err error) error {
	//		if errors.Is(err, fiber.ErrNotFound) {
	//			return ctx.Status(fiber.StatusNotFound).SendString(err.Error())
	//		}
	//		return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	//	},
	//})
	app := fiber.New()

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(healthcheck.New())

	RegisterRoutes(app)

	log.Fatal(app.Listen(":1323"))
}

func initialize() {
	MakeDir()
}
