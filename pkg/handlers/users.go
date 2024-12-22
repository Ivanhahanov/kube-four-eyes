package handlers

import (
	"webhook/pkg/coordination"
	"webhook/pkg/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func onlineStatus(users map[string]string, uid string) string {
	if val, ok := users[uid]; ok {
		return val
	}
	return "offline"
}

func UsersList(ctx *fiber.Ctx) error {
	rid := ctx.Params("id")
	statuses, err := coordination.GetStatuses(rid)
	if err != nil {
		log.Error("get statuses", err)
		return fiber.ErrBadRequest
	}
	onlineUsers, err := coordination.GetOnline(rid)
	if err != nil {
		log.Error("get statuses", err)
		return fiber.ErrBadRequest
	}

	var resp = []models.WebsocketMessage{}
	for _, user := range coordination.Policy().Users {

		resp = append(resp, models.WebsocketMessage{
			UserId:    user,
			RequestId: rid,
			Type:      "user_update",
			Active:    onlineStatus(onlineUsers, user),
			Status:    statuses[user],
		})
	}
	return ctx.JSON(resp)
}
