package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	. "leita/src/commands"
	. "leita/src/entities"
	"leita/src/services"
	. "leita/src/utils"
)

// JudgeProblem godoc
//
//	@Accept		json
//	@Produce	json
//	@Tags		Problem
//	@Param		problemId	path		string				true	"problemId"
//	@Param		requestBody	body		JudgeProblemRequest	true	"requestBody"
//	@Success	200			{object}	JudgeProblemResponse
//	@Failure	500			{object}	JudgeProblemResponse
//	@Router		/problem/{problemId} [post]
func JudgeProblem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req JudgeProblemRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(JudgeProblemResponse{
				IsSuccessful: false,
				Error:        "Invalid request body",
			})
		}

		problemId := c.Params("problemId")
		submitId := strconv.Itoa(req.SubmitId)
		language := req.Language
		code := Decode(req.Code)
		testcases := 1
		command := Commands[language]
		requireBuild := command.RequireBuild
		buildCmd := ReplaceSubmitId(command.BuildCmd, submitId)
		runCmd := ReplaceSubmitId(command.RunCmd, submitId)
		deleteCmd := ReplaceSubmitId(command.DeleteCmd, submitId)

		judgeProblemDTO := JudgeProblemDTO{
			ProblemId:    problemId,
			SubmitId:     submitId,
			Language:     language,
			Code:         code,
			Testcases:    testcases,
			RequireBuild: requireBuild,
			BuildCmd:     buildCmd,
			RunCmd:       runCmd,
			DeleteCmd:    deleteCmd,
		}

		problemService := services.NewProblemService()
		judgeProblemResult := problemService.JudgeProblem(judgeProblemDTO)

		if judgeProblemResult.Error != nil {
			return c.Status(judgeProblemResult.Status).JSON(JudgeProblemResponse{
				IsSuccessful: judgeProblemResult.IsSuccessful,
				Error:        judgeProblemResult.Error.Error(),
			})
		}

		return c.Status(judgeProblemResult.Status).JSON(JudgeProblemResponse{
			IsSuccessful: judgeProblemResult.IsSuccessful,
			Error:        "",
		})
	}
}
