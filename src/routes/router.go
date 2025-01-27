package routes

import (
	"github.com/gofiber/fiber/v3"
	"leita/src/apis/judge"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	judgeGroup := api.Group("/judge")
	judgeGroup.Post("", judge.JudgeProblem)
}
