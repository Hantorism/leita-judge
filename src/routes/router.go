package routes

import (
	"github.com/gofiber/fiber/v3"
	"leita/src/apis/problem"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	problemGroup := api.Group("/problem")
	problemGroup.Post("/:problemId", problem.JudgeProblem)
}
