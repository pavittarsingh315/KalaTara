package authcontrollers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rivo/uniseg"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/configs/cache"
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
	cacheCtx, cacheCancel := cache.NewCacheContext()
	defer cacheCancel()
	var key = cache.PasswordResetCodeKey(reqBody.Contact)
	var resetCode string
	if err := cache.Get(cacheCtx, key, &resetCode); err == nil { // no error => key exists ie hasnt expired
		cacheCtx, cacheCancel := cache.NewCacheContext()
		defer cacheCancel()
		dur, _ := cache.ExpiresIn(cacheCtx, key)
		message := fmt.Sprintf("Try again in %s.", utils.SecondsToString(int64(dur.Seconds())))
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": message}))
	} else if err != redis.Nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Create password reset code in cache
	cacheCtx2, cacheCancel2 := cache.NewCacheContext()
	defer cacheCancel2()
	var code = utils.GenerateRandomCode(6)
	var exp = cache.PasswordResetCodeEXP
	if err := cache.Set(cacheCtx2, key, utils.HashPassword(code), exp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Send Reset Code
	contactIsEmail := utils.ValidateEmail(reqBody.Contact)
	if contactIsEmail {
		go utils.SendPasswordResetEmail(user.Name, user.Contact, code)
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "An email has been sent with a reset verification code."}))
	} else {
		go utils.SendPasswordResetText(code, user.Contact)
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
	cacheCtx, cacheCancel := cache.NewCacheContext()
	defer cacheCancel()
	var key = cache.PasswordResetCodeKey(reqBody.Contact)
	var resetCode string
	if err := cache.Get(cacheCtx, key, &resetCode); err != nil {
		if err == redis.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Code has expired. Please restart the recovery process."}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
		}
	}

	// Check if user provided code is correct
	if !utils.VerifyPassword(resetCode, reqBody.Code) {
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
	cacheCtx, cacheCancel := cache.NewCacheContext()
	defer cacheCancel()
	var key = cache.PasswordResetCodeKey(reqBody.Contact)
	var resetCode string
	if err := cache.Get(cacheCtx, key, &resetCode); err != nil {
		if err == redis.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Code has expired. Please restart the recovery process."}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
		}
	}

	// Check if user provided code is correct
	if !utils.VerifyPassword(resetCode, reqBody.Code) {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Incorrect Code."}))
	}

	// We know that a reset code is created after checking if the user with contact = reqBody.Contact exists.
	// So theres no need to check now if a user exists with contact = reqBody.Contact because the only way the reset code is created is if thats true.

	// Update password
	if err := configs.Database.Model(&models.User{}).Where("contact = ?", reqBody.Contact).Update("password", utils.HashPassword(reqBody.Password)).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Delete reset code from cache
	cacheCtx2, cacheCancel2 := cache.NewCacheContext()
	defer cacheCancel2()
	cache.Delete(cacheCtx2, key)

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Password has successfully been updated."}))
}
