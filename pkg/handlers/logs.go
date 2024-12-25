package handlers

import (
	"webhook/pkg/coordination"

	"github.com/gofiber/fiber/v2"
)

func GetLogs(c *fiber.Ctx) error {
	logs := coordination.GetLog()
	return c.JSON(fiber.Map{
		"status": "ok",
		"logs":   logs,
	})
}
