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
			return c.Status(fiber.StatusBadRequest).JSON(SubmitProblemResponse{
				Result: "",
				Error:  err.Error(),
			})
		}

		problemId, err := strconv.Atoi(c.Params("problemId"))
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(SubmitProblemResponse{
				Result: "",
				Error:  err.Error(),
			})
		}
		submitId := req.SubmitId
		language := req.Language
		code, err := Decode(req.Code)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(SubmitProblemResponse{
				Result: "",
				Error:  err.Error(),
			})
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

		result, err := handler.service.SubmitProblem(submitProblemDTO)
		if result == JudgeUnknown {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON(SubmitProblemResponse{
				Result: JudgeUnknown.String(),
				Error:  err.Error(),
			})
		}

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		return c.Status(fiber.StatusOK).JSON(SubmitProblemResponse{
			Result: result.String(),
			Error: errMsg,
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
			return c.Status(fiber.StatusBadRequest).JSON([]RunProblemResponse{
				{
					Result: "",
					Error:  err.Error(),
				},
			})
		}

		problemId, err := strconv.Atoi(c.Params("problemId"))
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON([]SubmitProblemResponse{
				{
					Result: "",
					Error:  err.Error(),
				},
			})
		}
		language := req.Language
		code, err := Decode(req.Code)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusInternalServerError).JSON([]RunProblemResponse{
				{
					Result: "",
					Error:  err.Error(),
				},
			})
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

		results := handler.service.RunProblem(runProblemDTO)

		responses := make([]RunProblemResponse, 0, len(results))
		for _, result := range results {
			errMsg := ""
			if result.Error != nil {
				errMsg = result.Error.Error()
			}
			responses = append(responses, RunProblemResponse{
				Result: result.Result.String(),
				Error: errMsg,
			})
		}

		return c.Status(fiber.StatusOK).JSON(responses)
	}
}
