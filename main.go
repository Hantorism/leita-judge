package main

import (
	"os"

	"leita/src/route"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title			Leita API Docs
// @version		2.0.0
// @description	Leita Judge System API Documentation
// @BasePath		/api
func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Info(err)
	}

	app := fiber.New()

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New())

	app.Get("/swagger.json", static.New("./docs/swagger.json"))
	app.Get("/api/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
	))

	if err := route.RegisterRoutes(app); err != nil {
		log.Fatal(err)
		return
	}

	log.Fatal(app.Listen(":" + os.Getenv("JUDGE_PORT")))
}
