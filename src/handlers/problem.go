package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	. "leita/src/commands"
	. "leita/src/entities"
	"leita/src/services"
	. "leita/src/utils"
)

// SubmitProblem godoc
//
//	@Accept		json
//	@Produce	json
//	@Tags		Problem
//	@Param		problemId	path		string				true	"problemId"
//	@Param		requestBody	body		SubmitProblemRequest	true	"requestBody"
//	@Success	200			{object}	SubmitProblemResponse
//	@Failure	500			{object}	SubmitProblemResponse
//	@Router		/problem/{problemId} [post]
func SubmitProblem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req SubmitProblemRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(SubmitProblemResponse{
				IsSuccessful: false,
				Error:        err.Error(),
			})
		}

		problemId, _ := strconv.Atoi(c.Params("problemId"))
		submitId := req.SubmitId
		language := req.Language
		code := Decode(req.Code)
		testcases := 1
		command := Commands[language]
		requireBuild := command.RequireBuild
		buildCmd := ReplaceSubmitId(command.BuildCmd, submitId)
		runCmd := ReplaceSubmitId(command.RunCmd, submitId)
		deleteCmd := ReplaceSubmitId(command.DeleteCmd, submitId)

		submitProblemDTO := SubmitProblemDTO{
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
		submitProblemResult := problemService.SubmitProblem(submitProblemDTO)

		if submitProblemResult.Error != nil {
			return c.Status(submitProblemResult.Status).JSON(SubmitProblemResponse{
				IsSuccessful: submitProblemResult.IsSuccessful,
				Error:        submitProblemResult.Error.Error(),
			})
		}

		return c.Status(submitProblemResult.Status).JSON(SubmitProblemResponse{
			IsSuccessful: submitProblemResult.IsSuccessful,
			Error:        "",
		})
	}
}
