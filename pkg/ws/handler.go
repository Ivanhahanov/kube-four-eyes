package ws

import (
	"encoding/json"
	"webhook/pkg/auth"
	"webhook/pkg/coordination"
	"webhook/pkg/models"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
)

func getUserId(c *websocket.Conn) string {
	return c.Locals("claims").(*auth.Claims).Email
}

func switchReq(rid, uid string) string {
	msg, _ := json.Marshal(models.WebsocketMessage{
		RequestId: rid,
		UserId:    uid,
		Type:      "switch_req",
	})
	return string(msg)
}

func Websocket(c *websocket.Conn) {
	// When the function returns, unregister the client and close the connection
	defer func() {
		Ch.Unregister <- c
		coordination.SetOffline(c.Params("id"), getUserId(c))
		c.Close()
	}()

	// Register the client
	Ch.Register <- c
	coordination.SetOnline(c.Params("id"), getUserId(c))
	Ch.Broadcast <- switchReq(c.Params("id"), getUserId(c))
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Info("read error:", err)
			}
			// disconnected
			return
		}
	}
}
