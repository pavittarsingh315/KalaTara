package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/configs/cache"
	"nerajima.com/NeraJima/routes"
	"nerajima.com/NeraJima/ws"
)

func main() {
	configs.InitEnv()

	app := fiber.New()

	// Middleware
	app.Use(helmet.New())
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())
	if !configs.EnvProdActive() {
		app.Get("/metrics", monitor.New(monitor.ConfigDefault))
	}

	configs.InitDatabase()
	cache.Initialize()

	hub := ws.NewHub()
	go hub.Run()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("ws-hub", hub)
		return c.Next()
	})
	routes.InitRouter(app, hub)

	// Launch Application
	if err := app.Listen(":" + os.Getenv("PORT")); err != nil {
		log.Fatalf("ERROR: app failed to start: %v", err)
	}
}
