package handlers

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	. "leita/src/commands"
	. "leita/src/entities"
	"leita/src/services"
	. "leita/src/utils"
)

type ProblemHandler interface {
	SubmitProblem() fiber.Handler
	RunProblem() fiber.Handler
}

type problemHandler struct {
	service services.ProblemService
}

func NewProblemHandler() (ProblemHandler, error) {
	service, err := services.NewProblemService()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &problemHandler{
		service: service,
	}, nil
}

// SubmitProblem godoc
//
//	@Accept		json
//	@Produce	json
//	@Tags		Problem
//	@Param		problemId	path		string					true	"problemId"
//	@Param		requestBody	body		SubmitProblemRequest	true	"requestBody"
//	@Success	200			{object}	SubmitProblemResponse
//	@Failure	500			{object}	SubmitProblemResponse
//	@Router		/problem/submit/{problemId} [post]
func (handler *problemHandler) SubmitProblem() fiber.Handler {
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
		code, err := Decode(req.Code)
		if err != nil {
			log.Error(err)
			return err
		}
		command := Commands[language]
		buildCmd := ReplaceCommand(command.BuildCmd, "submit", submitId)
		runCmd := ReplaceCommand(command.RunCmd, "submit", submitId)
		deleteCmd := ReplaceCommand(command.DeleteCmd, "submit", submitId)

		submitProblemDTO := SubmitProblemDTO{
			ProblemId: problemId,
			SubmitId:  submitId,
			Language:  language,
			Code:      code,
			BuildCmd:  buildCmd,
			RunCmd:    runCmd,
			DeleteCmd: deleteCmd,
		}

		submitProblemResult, err := handler.service.SubmitProblem(submitProblemDTO)
		if err != nil {
			log.Error(err)
			return err
		}

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

// RunProblem godoc
//
//	@Accept		json
//	@Produce	json
//	@Tags		Problem
//	@Param		problemId	path		string				true	"problemId"
//	@Param		requestBody	body		RunProblemRequest	true	"requestBody"
//	@Success	200			{object}	RunProblemResponse
//	@Failure	500			{object}	RunProblemResponse
//	@Router		/problem/run/{problemId} [post]
func (handler *problemHandler) RunProblem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RunProblemRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(RunProblemResponse{
				IsSuccessful: []bool{},
				Error:        err.Error(),
			})
		}

		problemId, _ := strconv.Atoi(c.Params("problemId"))
		language := req.Language
		code, err := Decode(req.Code)
		if err != nil {
			log.Error(err)
			return err
		}
		testCases := req.TestCases
		submitId := RandomInt(int(math.Pow10(11)), int(math.Pow10(12)-1))
		command := Commands[language]
		buildCmd := ReplaceCommand(command.BuildCmd, "run", submitId)
		runCmd := ReplaceCommand(command.RunCmd, "run", submitId)
		deleteCmd := ReplaceCommand(command.DeleteCmd, "run", submitId)

		runProblemDTO := RunProblemDTO{
			ProblemId: problemId,
			SubmitId:  submitId,
			Language:  language,
			Code:      code,
			TestCases: testCases,
			BuildCmd:  buildCmd,
			RunCmd:    runCmd,
			DeleteCmd: deleteCmd,
		}

		runProblemResult, err := handler.service.RunProblem(runProblemDTO)
		if err != nil {
			log.Error(err)
			return err
		}

		if runProblemResult.Error != nil {
			return c.Status(runProblemResult.Status).JSON(RunProblemResponse{
				IsSuccessful: runProblemResult.IsSuccessful,
				Error:        runProblemResult.Error.Error(),
			})
		}

		return c.Status(runProblemResult.Status).JSON(RunProblemResponse{
			IsSuccessful: runProblemResult.IsSuccessful,
			Error:        "",
		})
	}
}
