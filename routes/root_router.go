package routes

import (
	"nerajima.com/NeraJima/responses"
	"nerajima.com/NeraJima/ws"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func InitRouter(app *fiber.App, hub *ws.Hub) {
	api := app.Group("/api")
	ws := app.Group("/ws")

	api.Get("/default", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("🚀🚀🚀🚀 - PSJ 11-04-22 6:56 pm")
	})

	AuthRouter(api)
	ProfileRouter(api)
	PostsRouter(api)

	ws.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) { // Returns true if the client requested upgrade to the WebSocket protocol
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})
	WSRouter(ws, hub)

	app.Use("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.NewErrorResponse(
				fiber.StatusNotFound,
				&fiber.Map{
					"data": "404 not found.",
				},
				nil,
			),
		)
	})
}
