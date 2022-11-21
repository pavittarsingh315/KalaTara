package profilecontrollers

import (
	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func EditUsername(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": reqProfile}))
}

func EditName(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Edit Name"}))
}

func EditBio(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Edit Bio"}))
}

func EditAvatar(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Edit Avatar"}))
}
