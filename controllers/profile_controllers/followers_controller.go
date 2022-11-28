package profilecontrollers

import (
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
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
	}
	if toBeFollowedProfile.Id == "" { // Id field is empty => user does not exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "The user you are trying to follow does not exist."}))
	}

	// Check if we already follow the user
	var followerObj models.ProfileFollower
	if err := configs.Database.Table("profile_followers").Find(&followerObj, "profile_id = ? AND follower_id = ?", toBeFollowedProfile.Id, reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
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
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
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
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
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
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Follower has been removed."}))
}

func GetFollowers(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	var profile models.Profile
	if err := configs.Database.Model(&models.Profile{}).Find(&profile, "id = ?", c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
	}
	if profile.Id == "" { // Id field is empty => user does not exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This user does not exist."}))
	}

	// Get followers(paginated)
	var followers []models.MiniProfile
	if err := configs.Database.Model(&profile).Offset(offset).Limit(limit).Association("Followers").Find(&followers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected error..."}))
	}

	// Get total number of followers
	numFollowers := configs.Database.Model(&profile).Association("Followers").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numFollowers) / float64(limit))),
			"data":         followers,
		},
	}))
}

func GetFollowing(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get following"}))
}
