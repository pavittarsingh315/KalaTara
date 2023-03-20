package routes

import (
	"nerajima.com/NeraJima/responses"

	"github.com/gofiber/fiber/v2"
)

func InitRouter(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/default", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("🚀🚀🚀🚀 - PSJ 11-04-22 6:56 pm")
	})

	AuthRouter(api)
	ProfileRouter(api)
	PostsRouter(api)

	app.Use("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.NewErrorResponse(
				fiber.StatusNotFound,
				&fiber.Map{
					"data": "404 not found.",
				},
			),
		)
	})
}
