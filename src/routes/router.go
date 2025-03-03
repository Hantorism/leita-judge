package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func RegisterRoutes(app *fiber.App) error {
	api := app.Group("/api")

	if err := RegisterProblemRoutes(api); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
