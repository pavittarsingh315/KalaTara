package authcontrollers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/configs/cache"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
	"nerajima.com/NeraJima/utils"
)

func Login(c *fiber.Ctx) error {
	reqBody := struct {
		Contact  string `json:"contact"`
		Password string `json:"password"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}, err))
	}

	// Check if all fields are included
	if reqBody.Contact == "" || reqBody.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}, nil))
	}

	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", "")) // remove all whitespace and make lowercase

	// Check if user exists
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var user models.User
	if err := configs.Database.WithContext(dbCtx).Model(&models.User{}).Preload("Profile").Find(&user, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	if user.Contact == "" || user.Profile.Username == "" { // (contact field is empty => user doesn't exist || username field is empty => profile doesn't exist) => Account is not found
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Account not found."}, nil))
	}

	if !utils.VerifyPassword(user.Password, reqBody.Password) { // password doesn't match
		return c.Status(fiber.StatusUnauthorized).JSON(responses.NewErrorResponse(fiber.StatusUnauthorized, &fiber.Map{"data": "Incorrect Password."}, nil))
	}

	// Check if user is banned
	unixTimeNow := time.Now().Unix()
	unixTimeBan := user.BanTill.Unix()
	if unixTimeNow < unixTimeBan {
		message := fmt.Sprintf("You are banned for %s.", utils.SecondsToString(unixTimeBan-unixTimeNow))
		return c.Status(fiber.StatusUnauthorized).JSON(responses.NewErrorResponse(fiber.StatusUnauthorized, &fiber.Map{"data": message}, nil))
	}

	// Generate auth tokens
	access, refresh := utils.GenAuthTokens(user.Id)

	go func() {
		// Update last login - because we preloaded the profile in the earlier query, we need to create a query on a "clean" user model so that a profile's username unique constraint isn't violated.
		dbCtx2, dbCancel2 := configs.NewQueryContext()
		defer dbCancel2()
		_ = configs.Database.WithContext(dbCtx2).Model(&models.User{}).Where("id = ?", user.Id).Update("last_login", time.Now()).Error

		// Cache profile
		cacheCtx, cacheCancel := cache.NewCacheContext()
		defer cacheCancel()
		var key = cache.ProfileKey(user.Id)
		var exp = cache.ProfileExp
		_ = cache.Set(cacheCtx, key, user.Profile, exp)
	}()

	return c.Status(fiber.StatusOK).JSON(
		responses.NewSuccessResponse(
			fiber.StatusOK,
			&fiber.Map{
				"data": &fiber.Map{
					"access":  access,
					"refresh": refresh,
					"profile": user.Profile,
				},
			},
		),
	)
}

func TokenLogin(c *fiber.Ctx) error {
	reqBody := struct {
		AccessToken  string `json:"access"`
		RefreshToken string `json:"refresh"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}, err))
	}

	// Check if all fields are included
	if reqBody.AccessToken == "" || reqBody.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}, nil))
	}

	// Verify tokens
	accessToken, accessBody, accessErr := utils.VerifyAccessToken(reqBody.AccessToken)
	_, refreshBody, refreshErr := utils.VerifyRefreshToken(reqBody.RefreshToken)
	if accessErr != nil || refreshErr != nil {
		err := accessErr
		if accessErr == nil {
			err = refreshErr
		}
		return c.Status(fiber.StatusUnauthorized).JSON(responses.NewErrorResponse(fiber.StatusUnauthorized, &fiber.Map{"data": "Authentication failed..."}, err))
	}

	if accessBody.UserId != refreshBody.UserId { // token pair are a mismatch
		return c.Status(fiber.StatusUnauthorized).JSON(responses.NewErrorResponse(fiber.StatusUnauthorized, &fiber.Map{"data": "Authentication failed..."}, nil))
	}

	// Update token in case verification process updated it
	reqBody.AccessToken = accessToken

	// Check if user exists
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var user models.User
	if err := configs.Database.WithContext(dbCtx).Model(&models.User{}).Preload("Profile").Find(&user, "id = ?", accessBody.UserId).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	if user.Contact == "" || user.Profile.Username == "" { // (contact field is empty => user doesn't exist || username field is empty => profile doesn't exist) => Account is not found
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Account not found."}, nil))
	}

	// Check if user is banned
	unixTimeNow := time.Now().Unix()
	unixTimeBan := user.BanTill.Unix()
	if unixTimeNow < unixTimeBan {
		message := fmt.Sprintf("You are banned for %s.", utils.SecondsToString(unixTimeBan-unixTimeNow))
		return c.Status(fiber.StatusUnauthorized).JSON(responses.NewErrorResponse(fiber.StatusUnauthorized, &fiber.Map{"data": message}, nil))
	}

	go func() {
		// Update last login - because we preloaded the profile in the earlier query, we need to create a query on a "clean" user model so that a profile's username unique constraint isn't violated.
		dbCtx2, dbCancel2 := configs.NewQueryContext()
		defer dbCancel2()
		_ = configs.Database.WithContext(dbCtx2).Model(&models.User{}).Where("id = ?", user.Id).Update("last_login", time.Now()).Error

		// Cache profile
		cacheCtx, cacheCancel := cache.NewCacheContext()
		defer cacheCancel()
		var key = cache.ProfileKey(user.Id)
		var exp = cache.ProfileExp
		_ = cache.Set(cacheCtx, key, user.Profile, exp)
	}()

	return c.Status(fiber.StatusOK).JSON(
		responses.NewSuccessResponse(
			fiber.StatusOK,
			&fiber.Map{
				"data": &fiber.Map{
					"access":  reqBody.AccessToken,
					"refresh": reqBody.RefreshToken,
					"profile": user.Profile,
				},
			},
		),
	)
}
