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
	"nerajima.com/NeraJima/routes"
)

func main() {
	configs.InitEnv()

	app := fiber.New()

	// Middleware
	app.Use(helmet.New())
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())
	if os.Getenv("APP_ENV") == "development" {
		app.Get("/metrics", monitor.New(monitor.ConfigDefault))
	}

	configs.InitDatabase()
	routes.InitRouter(app)

	// Launch Application
	if err := app.Listen(":" + os.Getenv("PORT")); err != nil {
		log.Fatal("ERROR: app failed to start")
		panic(err)
	}
}
