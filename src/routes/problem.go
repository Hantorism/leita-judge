package routes

import (
	"github.com/gofiber/fiber/v2"
	"leita/src/handlers"
)

func RegisterProblemRoutes(api fiber.Router) {
	handler := handlers.NewProblemHandler()

	problemGroup := api.Group("/problem")
	problemGroup.Post("/submit/:problemId", handler.SubmitProblem())
}
