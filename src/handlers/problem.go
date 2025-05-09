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

type ProblemHandler struct {
	service *services.ProblemService
}

func NewProblemHandler() (*ProblemHandler, error) {
	service, err := services.NewProblemService()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &ProblemHandler{
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
func (handler *ProblemHandler) SubmitProblem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req SubmitProblemRequest
		if err := c.BodyParser(&req); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(SubmitProblemResponse{
				Error: err.Error(),
			})
		}

		problemId, _ := strconv.Atoi(c.Params("problemId"))
		submitId := req.SubmitId
		language := req.Language
		code := DecodeBase64([]byte(req.Code))
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

		result, usedTime, usedMemory, err := handler.service.SubmitProblem(submitProblemDTO)
		if result == JudgeUnknown {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(SubmitProblemResponse{
				Result: JudgeUnknown.String(),
				Error:  err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(SubmitProblemResponse{
			Result:     result.String(),
			Error: ErrStrIfNotNil(err),
			UsedTime:   usedTime,
			UsedMemory: usedMemory,
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
//	@Success	200			{object}	[]RunProblemResponse
//	@Failure	500			{object}	[]RunProblemResponse
//	@Router		/problem/run/{problemId} [post]
func (handler *ProblemHandler) RunProblem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RunProblemRequest
		if err := c.BodyParser(&req); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON([]RunProblemResponse{
				{
					Error: err.Error(),
				},
			})
		}

		problemId, _ := strconv.Atoi(c.Params("problemId"))
		language := req.Language
		code := DecodeBase64([]byte(req.Code))
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

		results := handler.service.RunProblem(runProblemDTO)

		responses := make([]RunProblemResponse, 0, len(results))
		for _, result := range results {
			responses = append(responses, RunProblemResponse{
				Result: result.Result.String(),
				Error: ErrStrIfNotNil(result.Error),
				Output: result.Output,
			})
		}

		return c.Status(fiber.StatusOK).JSON(responses)
	}
}
