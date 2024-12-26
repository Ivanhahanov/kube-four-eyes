package main

import (
	"fmt"
	"os"
	"webhook/pkg/auth"
	"webhook/pkg/handlers"
	"webhook/pkg/helpers"
	"webhook/pkg/ws"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	var validator func(*fiber.Ctx, string) (bool, error)
	var errorHandler func(*fiber.Ctx, error) error
	var next func(*fiber.Ctx) bool
	var err error
	switch helpers.GetEnv("AUTH_TYPE", "dex") {
	case "dex":
		validator, err = auth.NewJWTValidator(
			helpers.GetEnv("OIDC_URL", "http://localhost:8080/realms/master"),
			helpers.GetEnv("OIDC_CLIENT_ID", "oauth2-proxy"),
		)
		if err != nil {
			log.Info(err)
		}
		// errorHandler = func(c *fiber.Ctx, err error) error {
		// 	c.ClearCookie()
		// 	return c.Redirect(helpers.GetEnv("AUTH_REDIRECT", "https://google.com"))
		// }
	case "fake":
		validator = func(c *fiber.Ctx, s string) (bool, error) {
			log.Debug("requested key", s)
			return true, nil
		}
		next = func(c *fiber.Ctx) bool {
			// request
			c.Request().Header.VisitAll(func(key, value []byte) {
				log.Debug("req headerKey", string(key), "value", string(value))
			})

			c.Request().Header.VisitAllCookie(func(key, value []byte) {
				log.Debug("req cookieKey", string(key), "value", string(value))
			})
			c.Locals("claims", &auth.Claims{
				Name:  "User",
				Email: "user1@example.com",
			})
			return true
		}
	}

	app := fiber.New(fiber.Config{ReadBufferSize: 8192})
	auth := keyauth.New(keyauth.Config{
		Validator:    validator,
		ErrorHandler: errorHandler,
		Next:         next,
		// KeyLookup: "cookie:_oauth2_proxy",
	})
	app.Use(logger.New())
	app.Static("/", fmt.Sprintf("%s/index.html", os.Getenv("KO_DATA_PATH")))
	app.Static("req/", fmt.Sprintf("%s/form.html", os.Getenv("KO_DATA_PATH")))
	app.Post("/authorize", handlers.Authorize)
	app.Get("/healthz", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(200)
	})

	go ws.Ch.RunHub()

	api := app.Group("/api", auth)
	api.Post("/submit-access-request", auth, handlers.SubmitAccessRequest)
	api.Get("/requests", auth, handlers.AccessRequestList)
	api.Get("/logs", auth, handlers.GetLogs)

	api.Use("/ws/:id", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired

	})
	api.Get("/ws/:id", websocket.New(ws.Websocket))
	api.Get("/users/:id", handlers.UsersList)
	api.Post("/approve/:id", handlers.Approve)
	api.Post("/ready/:id", handlers.Ready)
	if err := app.Listen(":9443"); err != nil {
		log.Fatal(err)
	}
}
