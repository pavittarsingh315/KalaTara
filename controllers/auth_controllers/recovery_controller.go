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
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}, err))
	}

	// Check if all fields are included
	if reqBody.Contact == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}, nil))
	}

	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", "")) // remove all whitespace and make lowercase

	// Check if account exists
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var user models.User
	if err := configs.Database.WithContext(dbCtx).Model(&models.User{}).Find(&user, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	if user.Contact == "" { // contact field is empty => user with contact doesn't exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Account not found."}, nil))
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
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": message}, nil))
	} else if err != redis.Nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Create password reset code in cache
	cacheCtx2, cacheCancel2 := cache.NewCacheContext()
	defer cacheCancel2()
	var code, err = utils.GenerateRandomCode(6)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	var hash, err2 = utils.HashPassword(code)
	if err2 != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err2))
	}
	var exp = cache.PasswordResetCodeEXP
	if err := cache.Set(cacheCtx2, key, hash, exp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
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
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}, err))
	}

	// Check if all fields are included
	if reqBody.Contact == "" || reqBody.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}, nil))
	}

	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", "")) // remove all whitespace and make lowercase

	// Check if reset code exists
	cacheCtx, cacheCancel := cache.NewCacheContext()
	defer cacheCancel()
	var key = cache.PasswordResetCodeKey(reqBody.Contact)
	var resetCode string
	if err := cache.Get(cacheCtx, key, &resetCode); err != nil {
		if err == redis.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Code has expired. Please restart the recovery process."}, nil))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
		}
	}

	// Check if user provided code is correct
	if !utils.VerifyPassword(resetCode, reqBody.Code) {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Incorrect Code."}, nil))
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
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}, err))
	}

	// Check if all fields are included
	if reqBody.Contact == "" || reqBody.Code == "" || reqBody.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}, nil))
	}

	passwordLength := uniseg.GraphemeClusterCount(reqBody.Password)
	if passwordLength < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Password too short."}, nil))
	}
	if length := len(reqBody.Password); length > 64 { // Since the max length password supported by bcrypt is 72 bytes, we check the length of the string in bytes. I made max length 64 to be safe rather than 72.
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Password too long."}, nil))
	}

	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", "")) // remove all whitespace and make lowercase

	// Check if reset code exists
	cacheCtx, cacheCancel := cache.NewCacheContext()
	defer cacheCancel()
	var key = cache.PasswordResetCodeKey(reqBody.Contact)
	var resetCode string
	if err := cache.Get(cacheCtx, key, &resetCode); err != nil {
		if err == redis.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Code has expired. Please restart the recovery process."}, nil))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
		}
	}

	// Check if user provided code is correct
	if !utils.VerifyPassword(resetCode, reqBody.Code) {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Incorrect Code."}, nil))
	}

	// We know that a reset code is created after checking if the user with contact = reqBody.Contact exists.
	// So theres no need to check now if a user exists with contact = reqBody.Contact because the only way the reset code is created is if thats true.

	// Update password
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var hash, err = utils.HashPassword(reqBody.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	if err := configs.Database.WithContext(dbCtx).Model(&models.User{}).Where("contact = ?", reqBody.Contact).Update("password", hash).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Delete reset code from cache
	cacheCtx2, cacheCancel2 := cache.NewCacheContext()
	defer cacheCancel2()
	cache.Delete(cacheCtx2, key)

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Password has successfully been updated."}))
}
