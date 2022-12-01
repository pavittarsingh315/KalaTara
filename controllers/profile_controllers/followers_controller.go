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
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot follow yourself."}))
	}

	var toBeFollowedProfile models.Profile
	if err := configs.Database.Model(&models.Profile{}).Find(&toBeFollowedProfile, "id = ?", c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if toBeFollowedProfile.Id == "" { // Id field is empty => user does not exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "The user you are trying to follow does not exist."}))
	}

	// Check if we already follow the user
	var followerObj models.ProfileFollower
	if err := configs.Database.Table("profile_followers").Find(&followerObj, "profile_id = ? AND follower_id = ?", toBeFollowedProfile.Id, reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if followerObj.ProfileId != "" && followerObj.FollowerId != "" { // if both fields are populated => reqUser is already following this user
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You're already following this user."}))
	}

	newFollowerObj := models.ProfileFollower{
		ProfileId:  toBeFollowedProfile.Id,
		FollowerId: reqProfile.Id,
		CreatedAt:  time.Now(),
	}
	if err := configs.Database.Table("profile_followers").Create(&newFollowerObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "User has been followed."}))
}

func UnfollowAUser(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot unfollow yourself."}))
	}

	// Delete followers object
	var followerObj models.ProfileFollower
	if err := configs.Database.Table("profile_followers").Delete(&followerObj, "profile_id = ? AND follower_id = ?", c.Params("profileId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "User has been unfollowed."}))
}

func RemoveAFollower(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot remove yourself."}))
	}

	// Delete followers object
	var followerObj models.ProfileFollower
	if err := configs.Database.Table("profile_followers").Delete(&followerObj, "profile_id = ? AND follower_id = ?", reqProfile.Id, c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Follower has been removed."}))
}

func GetFollowers(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	var profile models.Profile
	if err := configs.Database.Model(&models.Profile{}).Find(&profile, "id = ?", c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if profile.Id == "" { // Id field is empty => user does not exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This user does not exist."}))
	}

	// Get followers(paginated)
	regexMatch := fmt.Sprintf("%%%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	var followers []models.MiniProfile
	if err := configs.Database.Model(&profile).Offset(offset).Limit(limit).Order("profile_followers.created_at DESC").Where("username LIKE ? OR name LIKE ?", regexMatch, regexMatch).Association("Followers").Find(&followers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of followers
	numFollowers := configs.Database.Model(&profile).Where("username LIKE ? OR name LIKE ?", regexMatch, regexMatch).Association("Followers").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numFollowers) / float64(limit))),
			"data":         followers,
		},
	}))
}

func GetFollowing(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	var profile models.Profile
	if err := configs.Database.Model(&models.Profile{}).Find(&profile, "id = ?", c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if profile.Id == "" { // Id field is empty => user does not exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This user does not exist."}))
	}

	/*
	   IMPORTANT:
	      To build the raw queries below, I used the default gorm logger in Info mode to inspect the queries made by the GetFollowers endpoint when gettings the followers list and numfollowers.
	      I then used those queries and altered them to build the queries seen below.
	*/

	// Get following(paginated)
	regexMatch := fmt.Sprintf("%%%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := fmt.Sprintf("SELECT profiles.id, profiles.created_at, profiles.updated_at, profiles.user_id, profiles.username, profiles.name, profiles.bio, profiles.avatar, profiles.mini_avatar, profiles.birthday FROM profiles JOIN profile_followers ON profile_followers.follower_id = \"%s\" AND profile_followers.profile_id = profiles.id WHERE username LIKE \"%s\" OR name LIKE \"%s\" ORDER BY profile_followers.created_at DESC LIMIT %d OFFSET %d", c.Params("profileId"), regexMatch, regexMatch, limit, offset)
	var following = []models.MiniProfile{}
	if err := configs.Database.Raw(query).Scan(&following).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of following
	var numFollowing int
	query2 := fmt.Sprintf("SELECT count(*) FROM profiles JOIN profile_followers on profile_followers.follower_id = \"%s\" AND profile_followers.profile_id = profiles.id WHERE username LIKE \"%s\" OR name LIKE \"%s\"", c.Params("profileId"), regexMatch, regexMatch)
	if err := configs.Database.Raw(query2).Scan(&numFollowing).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numFollowing) / float64(limit))),
			"data":         following,
		},
	}))
}
