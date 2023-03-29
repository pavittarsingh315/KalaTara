package profilecontrollers

import (
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func FollowAUser(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot follow yourself."}, nil))
	}

	newFollowerObj := models.ProfileFollower{
		ProfileId:  c.Params("profileId"),
		FollowerId: reqProfile.Id,
		CreatedAt:  time.Now(),
	}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Table("profile_followers").Where("profile_id = ? AND follower_id = ?", c.Params("profileId"), reqProfile.Id).FirstOrCreate(&newFollowerObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "User has been followed."}))
}

func UnfollowAUser(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot unfollow yourself."}, nil))
	}

	// Delete followers object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var followerObj models.ProfileFollower
	if err := configs.Database.WithContext(dbCtx).Table("profile_followers").Delete(&followerObj, "profile_id = ? AND follower_id = ?", c.Params("profileId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "User has been unfollowed."}))
}

func RemoveAFollower(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot remove yourself."}, nil))
	}

	// Delete followers object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var followerObj models.ProfileFollower
	if err := configs.Database.WithContext(dbCtx).Table("profile_followers").Delete(&followerObj, "profile_id = ? AND follower_id = ?", reqProfile.Id, c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Follower has been removed."}))
}

func GetFollowers(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	// Get followers(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	var followers []models.MiniProfile
	if err := configs.Database.WithContext(dbCtx).Model(&models.Profile{Base: models.Base{Id: c.Params("profileId")}}).Offset(offset).Limit(limit).Order("profile_followers.created_at DESC").Where("username LIKE ? OR name LIKE ?", regexMatch, regexMatch).Association("Followers").Find(&followers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of followers
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	numFollowers := configs.Database.WithContext(dbCtx2).Model(&models.Profile{Base: models.Base{Id: c.Params("profileId")}}).Where("username LIKE ? OR name LIKE ?", regexMatch, regexMatch).Association("Followers").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numFollowers) / float64(limit))),
			"data":         followers,
		},
	}))
}

func GetFollowing(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	/*
	   IMPORTANT:
	      To build the raw queries below, I used the default gorm logger in Info mode to inspect the queries made by the GetFollowers endpoint when gettings the followers list and numfollowers.
	      I then used those queries and altered them to build the queries seen below.
	*/

	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := configs.Database.Table("profiles").
		Joins("JOIN profile_followers ON profile_followers.follower_id = ? AND profile_followers.profile_id = profiles.id", c.Params("profileId")).
		Where("username LIKE ? OR name LIKE ?", regexMatch, regexMatch)

		// Get following(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var following = []models.MiniProfile{}
	if err := query.WithContext(dbCtx).Order("profile_followers.created_at DESC").Limit(limit).Offset(offset).Scan(&following).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of following
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numFollowing int64
	if err := query.WithContext(dbCtx2).Count(&numFollowing).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numFollowing) / float64(limit))),
			"data":         following,
		},
	}))
}
