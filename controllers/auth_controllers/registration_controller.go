package authcontrollers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rivo/uniseg"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
	"nerajima.com/NeraJima/utils"
)

func InitiateRegistration(c *fiber.Ctx) error {
	reqBody := struct {
		Contact  string    `json:"contact"`
		Username string    `json:"username"`
		Name     string    `json:"name"`
		Password string    `json:"password"`
		Birthday time.Time `json:"birthday"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	// Check if all fields are included
	if reqBody.Contact == "" || reqBody.Username == "" || reqBody.Name == "" || reqBody.Password == "" || reqBody.Birthday.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}

	reqBody.Name = strings.TrimSpace(reqBody.Name)                                    // remove leading and trailing whitespace
	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", ""))   // remove all whitespace and make lowercase
	reqBody.Username = strings.ToLower(strings.ReplaceAll(reqBody.Username, " ", "")) // remove all whitespace and make lowercase

	// Validate request body lengths
	usernameLength := uniseg.GraphemeClusterCount(reqBody.Username)
	nameLength := uniseg.GraphemeClusterCount(reqBody.Name)
	contactLength := uniseg.GraphemeClusterCount(reqBody.Contact)
	passwordLength := uniseg.GraphemeClusterCount(reqBody.Password)
	if usernameLength < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username too short."}))
	}
	if usernameLength > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username too long."}))
	}
	if nameLength > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Name too long."}))
	}
	if contactLength > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Contact too long."}))
	}
	if passwordLength < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Password too short."}))
	}

	// Check if username is taken
	var profile models.Profile
	if err := configs.Database.Model(&models.Profile{}).Find(&profile, "username = ?", reqBody.Username).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
	}
	if profile.Username != "" { // username field is not empty => profile with username exists
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username is taken."}))
	}

	// Check if contact is taken
	contactIsEmail := utils.ValidateEmail(reqBody.Contact)
	var user models.User
	if err := configs.Database.Model(&models.User{}).Find(&user, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
	}
	if user.Contact != "" { // contact field is not empty => user with contact exists
		errorMsg := "Contact already in use."
		if contactIsEmail {
			errorMsg = "Email address already in use."
		}
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": errorMsg}))
	}

	// Check if registration is already initiated
	var tempObj models.TemporaryObject
	if err := configs.Database.Model(&models.TemporaryObject{}).Find(&tempObj, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
	}
	if tempObj.Contact != "" { // contact field is not empty => temporary object with contact exists
		unixTimeNow := time.Now().Unix()
		unixTimeFiveMinAfterObjCreated := tempObj.CreatedAt.Add(time.Minute * 5).Unix()
		if unixTimeFiveMinAfterObjCreated <= unixTimeNow { // tempObj is expired
			if err := configs.Database.Delete(&tempObj).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
			}
		} else { // tempObj is NOT expired
			message := fmt.Sprintf("Looks like this contact is part of another attempted registration. Try again in %s.", utils.SecondsToString(unixTimeFiveMinAfterObjCreated-unixTimeNow))
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": message}))
		}
	}

	// Create tempObj
	code := utils.EncodeToInt(6)
	newTempObj := models.TemporaryObject{
		VerificationCode: code,
		Contact:          reqBody.Contact,
	}
	if err := configs.Database.Create(&newTempObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
	}

	if contactIsEmail {
		// send email
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "An email has been sent with a verification code."}))
	} else {
		// send text
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "A text has been sent with a verification code."}))
	}
}

func FinalizeRegistration(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(
		responses.NewSuccessResponse(
			fiber.StatusOK,
			&fiber.Map{
				"data": "Finalize Register",
			},
		),
	)
}
