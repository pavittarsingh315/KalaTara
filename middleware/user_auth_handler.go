package middleware

import (
	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
	"nerajima.com/NeraJima/utils"
)

func UserAuthHandler(c *fiber.Ctx) error {
	reqHeader := struct {
		Token  string `reqHeader:"token"`
		UserId string `reqHeader:"userId"`
	}{}
	errMessage := "Could not authorize action."

	if err := c.ReqHeaderParser(&reqHeader); err != nil || reqHeader.Token == "" || reqHeader.UserId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.NewErrorResponse(fiber.StatusUnauthorized, &fiber.Map{"data": errMessage}))
	}

	_, accessBody, accessErr := utils.VerifyAccessTokenNoRefresh(reqHeader.Token) // will return err if expired
	if accessErr != nil || accessBody.UserId != reqHeader.UserId {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.NewErrorResponse(fiber.StatusUnauthorized, &fiber.Map{"data": errMessage}))
	}

	var user models.User
	if err := configs.Database.Model(&models.User{}).Preload("Profile").Find(&user, "id = ?", accessBody.UserId).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
	}
	if user.Id == "" || user.Profile.Username == "" { // (contact field is empty => user doesn't exist || username field is empty => profile doesn't exist) => Account is not found
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Account not found."}))
	}

	c.Locals("profile", user.Profile)

	return c.Next()
}
