package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/configs/cache"
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

	cacheCtx, cacheCancel := cache.NewCacheContext()
	defer cacheCancel()
	var profile models.Profile
	var key = cache.ProfileKey(accessBody.UserId)
	var exp = cache.ProfileExp
	if err := cache.Get(cacheCtx, key, &profile); err != nil {
		if err == redis.Nil { // key does not exist
			dbCtx, dbCancel := configs.NewQueryContext()
			defer dbCancel()
			if err := configs.Database.WithContext(dbCtx).Model(&models.Profile{}).Find(&profile, "user_id = ?", accessBody.UserId).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
			}
			if profile.Id == "" { // Id field is empty => Account is not found
				return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Account not found."}))
			}

			// Cache profile
			cacheCtx, cacheCancel := cache.NewCacheContext()
			defer cacheCancel()
			if err := cache.Set(cacheCtx, key, profile, exp); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
		}
	}

	c.Locals("profile", profile)

	return c.Next()
}
