package routes

import (
	"github.com/gofiber/fiber/v2"
	"leita/src/handlers"
	"leita/src/services"
)

func RegisterProblemRoutes(api fiber.Router) {
	problemGroup := api.Group("/problem")
	problemGroup.Post("/:problemId", handlers.JudgeProblem(services.NewProblemService()))
}
