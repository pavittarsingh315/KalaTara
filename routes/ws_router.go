package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"nerajima.com/NeraJima/middleware"
	"nerajima.com/NeraJima/ws"
)

func WSRouter(group fiber.Router, hub *ws.Hub) {
	router := group // domain/ws

	// TODO: look to move this connect logic into login routes where an extra redis call is not required
	router.Get("/connect", middleware.UserAuthHandler, websocket.New(hub.Connect)) // User should hit login route before this one so middleware shouldn't take too long since it'll use redis
}
