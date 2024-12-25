package handlers

import (
	"webhook/pkg/auth"
	"webhook/pkg/coordination"
	"webhook/pkg/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func SubmitAccessRequest(c *fiber.Ctx) error {
	var ar = new(models.AccessRequest)
	if err := c.BodyParser(ar); err != nil {
		log.Error(err)
		return fiber.ErrBadRequest
	}
	ar.Username = auth.GetUserId(c)
	ar.Email = auth.GetUserEmail(c)
	ar.TimePeriod += "h"
	ar.Cluster = "prod"
	log.Debug(ar)

	id, err := coordination.NewRequest(*ar)
	if err != nil {
		log.Error(err)
		return fiber.ErrBadRequest
	}
	return c.JSON(map[string]string{
		"status": "ok",
		"id":     string(id),
	})
}
