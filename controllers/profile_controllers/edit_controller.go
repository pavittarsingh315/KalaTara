package profilecontrollers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rivo/uniseg"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func EditUsername(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)
	reqBody := struct {
		Username string `json:"username"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	// Check if all fields are included
	if reqBody.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}

	reqBody.Username = strings.ToLower(strings.ReplaceAll(reqBody.Username, " ", "")) // remove all whitespace and make lowercase

	if reqBody.Username == reqProfile.Username {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This is your current username."}))
	}

	usernameLength := uniseg.GraphemeClusterCount(reqBody.Username)
	if usernameLength < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username is too short."}))
	}
	if usernameLength > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username is too long."}))
	}

	// Update username
	if err := configs.Database.Model(&reqProfile).Update("username", reqBody.Username).Error; err != nil {
		if err.Error() == "Error 1062: Duplicate entry 'darkstar' for key 'profiles.username'" {
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username is taken."}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
		}
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Username has been updated."}))
}

func EditName(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)
	reqBody := struct {
		Name string `json:"name"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	// no need to check if name is empty because its allowed to be empty.

	reqBody.Name = strings.TrimSpace(reqBody.Name) // remove leading and trailing whitespace

	if reqBody.Name == reqProfile.Name {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This is your current name."}))
	}

	nameLength := uniseg.GraphemeClusterCount(reqBody.Name)
	if nameLength > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Name is too long."}))
	}

	// Update name
	if err := configs.Database.Model(&reqProfile).Update("name", reqBody.Name).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Name has been updated."}))
}

func EditBio(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)
	reqBody := struct {
		Bio string `json:"bio"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	// no need to check if bio is empty because its allowed to be empty.

	reqBody.Bio = strings.TrimSpace(reqBody.Bio) // remove leading and trailing whitespace

	if reqBody.Bio == reqProfile.Bio {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This is your current bio."}))
	}

	bioLength := uniseg.GraphemeClusterCount(reqBody.Bio)
	if len(strings.Split(reqBody.Bio, "\n")) > 6 { // bio has 6 lines max
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Line limit exceeded."}))
	}
	if bioLength > 151 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bio is too long."}))
	}

	// Update bio
	if err := configs.Database.Model(&reqProfile).Update("bio", reqBody.Bio).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Bio has been updated."}))
}

func EditAvatar(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Edit Avatar"}))
}
