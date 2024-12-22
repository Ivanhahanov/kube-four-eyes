package handlers

import (
	"encoding/json"
	"webhook/pkg/auth"
	"webhook/pkg/coordination"
	"webhook/pkg/helpers"
	"webhook/pkg/models"
	"webhook/pkg/ws"

	"github.com/gofiber/fiber/v2"
)

func updateWsStatus(rid, uid, status string) error {
	msg, err := json.Marshal(models.WebsocketMessage{
		UserId:    uid,
		RequestId: rid,
		Type:      "update_status",
		Status:    status,
	})
	if err != nil {
		return err
	}
	ws.Ch.Broadcast <- string(msg)
	return nil
}

func Ready(ctx *fiber.Ctx) error {
	rid := ctx.Params("id")
	uid := auth.GetUserEmail(ctx)
	if uid != "" {
		coordination.ChangeStatus(rid, uid, helpers.StatusReady)
		updateWsStatus(rid, uid, helpers.StatusReady)
	}
	return ctx.JSON(map[string]string{
		"status": "ok",
	})
}

func Approve(ctx *fiber.Ctx) error {
	rid := ctx.Params("id")
	uid := auth.GetUserEmail(ctx)
	if uid == "" {
		return fiber.ErrBadRequest
	}

	coordination.ChangeStatus(rid, uid, helpers.StatusApproved)
	err := updateWsStatus(rid, uid, helpers.StatusReady)
	if err != nil {
		return fiber.ErrTeapot
	}
	if coordination.Policy().CheckPolicy(rid) {
		coordination.GrantUserAccess(rid)
	}
	return ctx.JSON(map[string]string{
		"status": "ok",
	})
}
