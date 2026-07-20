package problem

import (
	"encoding/base64"

	"leita/src/entity"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

var judgeSemaphore = make(chan struct{}, 4)

// SubmitProblem godoc
//
//	@Summary		Submit a problem solution
//	@Description	Submit code for a specific problem to be judged.
//	@Accept			json
//	@Produce		json
//	@Tags			Problem
//	@Param			problemId	path		string						true	"Problem ID"
//	@Param			requestBody	body		entity.SubmitProblemRequest	true	"Solution code and metadata"
//	@Success		200			{object}	entity.SubmitProblemReceiptResponse
//	@Failure		400			{object}	entity.SubmitProblemReceiptResponse
//	@Failure		500			{object}	entity.SubmitProblemReceiptResponse
//	@Router			/problem/submit/{problemId} [post]
func (handler *Handler) SubmitProblem() fiber.Handler {
	return func(c fiber.Ctx) error {
		var req entity.SubmitProblemRequest
		if err := c.Bind().Body(&req); err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(entity.SubmitProblemReceiptResponse{
				Status: "FAILED_BINDING",
			})
		}

		code, err := base64.StdEncoding.DecodeString(req.Code)
		if err != nil {
			log.Error(err)
			return c.Status(fiber.StatusBadRequest).JSON(entity.SubmitProblemReceiptResponse{
				Status: "FAILED_DECODING",
			})
		}

		dto := entity.SubmitProblemDTO{
			ProblemId: c.Params("problemId"),
			SubmitId:  req.SubmitId,
			Language:  req.Language,
			Code:      code,
			Limit:     req.Limit,
		}

		// 비동기로 고루틴 실행
		go func() {
			judgeSemaphore <- struct{}{}
			defer func() {
				<-judgeSemaphore
				if r := recover(); r != nil {
					log.Errorf("Panic in SubmitProblem goroutine for submit %d: %v", dto.SubmitId, r)
					_ = handler.service.PublishJudgeResult(dto.SubmitId, entity.JudgeUnknown, 0, 0, "Server Panic during judging")
				}
			}()

			result, usedTime, usedMemory, err := handler.service.SubmitProblem(dto)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			errPublish := handler.service.PublishJudgeResult(dto.SubmitId, result, usedTime, usedMemory, errStr)
			if errPublish != nil {
				log.Errorf("Failed to publish result to Redis for submit %d: %v", dto.SubmitId, errPublish)
			}
		}()

		return c.Status(fiber.StatusOK).JSON(entity.SubmitProblemReceiptResponse{
			Status: "RECEIVED",
		})
	}
}
