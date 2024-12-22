package handlers

import (
	"fmt"
	"webhook/pkg/coordination"
	"webhook/pkg/models"

	"github.com/gofiber/fiber/v2"
)

func CreateAccessRequest(ctx *fiber.Ctx) error {
	ar := new(models.AccessRequest)
	if err := ctx.BodyParser(ar); err != nil {
		return err
	}
	id, err := coordination.NewRequest(*ar)
	if err != nil {
		return ctx.JSON(fiber.Map{
			"status": "error",
			"err":    fmt.Errorf("can't create new access request: %e", err),
		})
	}
	return ctx.JSON(fiber.Map{
		"status": "ok",
		"id":     id,
	})
}

func AccessRequestList(ctx *fiber.Ctx) error {
	requests, err := coordination.GetAllRequests()
	if err != nil {
		return ctx.JSON(fiber.Map{
			"status": "error",
			"err":    fmt.Errorf("can't list access requests: %v", err).Error(),
		})
	}
	return ctx.JSON(fiber.Map{
		"status":   "ok",
		"requests": requests,
	})
}
