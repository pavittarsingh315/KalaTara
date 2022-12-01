package authcontrollers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rivo/uniseg"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
	"nerajima.com/NeraJima/utils"
)

func RequestPasswordReset(c *fiber.Ctx) error {
	reqBody := struct {
		Contact string `json:"contact"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	// Check if all fields are included
	if reqBody.Contact == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}

	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", "")) // remove all whitespace and make lowercase

	// Check if account exists
	var user models.User
	if err := configs.Database.Model(&models.User{}).Find(&user, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if user.Contact == "" { // contact field is empty => user with contact doesn't exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Account not found."}))
	}

	// Check if reset is already initiated
	var tempObj models.TemporaryObject
	if err := configs.Database.Model(&models.TemporaryObject{}).Find(&tempObj, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if tempObj.Contact != "" { // contact field is not empty => temporary object with contact exists
		if tempObj.IsExpired() {
			if err := configs.Database.Model(&models.TemporaryObject{}).Delete(&tempObj).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
			}
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please try again in a few minutes."}))
		}
	}

	// Create tempObj
	code := utils.GenerateRandomCode(6)
	newTempObj := models.TemporaryObject{
		VerificationCode: utils.HashPassword(code),
		Contact:          reqBody.Contact,
	}
	if err := configs.Database.Model(&models.TemporaryObject{}).Create(&newTempObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Send Reset Code
	contactIsEmail := utils.ValidateEmail(reqBody.Contact)
	if contactIsEmail {
		utils.SendPasswordResetEmail(user.Name, user.Contact, code)
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "An email has been sent with a reset verification code."}))
	} else {
		utils.SendPasswordResetText(code, user.Contact)
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "A text has been sent with a reset verification code."}))
	}
}

func ConfirmResetCode(c *fiber.Ctx) error {
	reqBody := struct {
		Contact string `json:"contact"`
		Code    string `json:"code"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	// Check if all fields are included
	if reqBody.Contact == "" || reqBody.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}

	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", "")) // remove all whitespace and make lowercase

	// Check if reset code exists
	var tempObj models.TemporaryObject
	if err := configs.Database.Model(&models.TemporaryObject{}).Find(&tempObj, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if tempObj.Contact == "" { // contact field is empty => temporary object with contact doesn't exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Code has expired. Please restart the reset process."}))
	} else {
		// If tempObj is expired, delete it
		if tempObj.IsExpired() {
			if err := configs.Database.Model(&models.TemporaryObject{}).Delete(&tempObj).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
			}
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Code has expired. Please restart the reset process."}))
		}
	}

	// Check if user provided code is correct
	if !utils.VerifyPassword(tempObj.VerificationCode, reqBody.Code) {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Incorrect Code."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Code confirmed."}))
}

func ConfirmPasswordReset(c *fiber.Ctx) error {
	reqBody := struct {
		Contact  string `json:"contact"`
		Password string `json:"password"`
		Code     string `json:"code"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	// Check if all fields are included
	if reqBody.Contact == "" || reqBody.Code == "" || reqBody.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}

	passwordLength := uniseg.GraphemeClusterCount(reqBody.Password)
	if passwordLength < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Password too short."}))
	}

	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", "")) // remove all whitespace and make lowercase

	// Check if reset code exists
	var tempObj models.TemporaryObject
	if err := configs.Database.Model(&models.TemporaryObject{}).Find(&tempObj, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if tempObj.Contact == "" { // contact field is empty => temporary object with contact doesn't exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Code has expired. Please restart the reset process."}))
	} else {
		// If tempObj is expired, delete it
		if tempObj.IsExpired() {
			if err := configs.Database.Model(&models.TemporaryObject{}).Delete(&tempObj).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
			}
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Code has expired. Please restart the reset process."}))
		}
	}

	// Check if user provided code is correct
	if !utils.VerifyPassword(tempObj.VerificationCode, reqBody.Code) {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Incorrect Code."}))
	}

	// Check if account exists
	var user models.User
	if err := configs.Database.Model(&models.User{}).Find(&user, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if user.Contact == "" { // contact field is empty => user with contact doesn't exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Account not found."}))
	}

	if utils.VerifyPassword(user.Password, reqBody.Password) { // old and new passwords match
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This is your current password...💀"}))
	}

	// Update password
	if err := configs.Database.Model(&user).Update("password", utils.HashPassword(reqBody.Password)).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Delete tempObj
	if err := configs.Database.Model(&models.TemporaryObject{}).Delete(&tempObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Password has successfully been updated."}))
}
