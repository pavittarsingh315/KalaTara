package postcontrollers

import (
	"math"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func GetFollowingsFeed(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get Followings Feed."}))
}

func GetSubscriptionsFeed(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get Subscriptions Feed."}))
}

func GetArchivedPosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Get archived posts(paginated)
	var archivedPosts = []models.Post{}
	if err := configs.Database.Model(&reqProfile).Offset(offset).Limit(limit).Order("posts.created_at DESC").Where("is_archived = ?", true).Preload("Media").Association("Posts").Find(&archivedPosts); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of archived posts
	numArchivedPosts := configs.Database.Model(&reqProfile).Where("is_archived = ?", true).Association("Posts").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numArchivedPosts) / float64(limit))),
			"data":         archivedPosts,
		},
	}))
}

func GetPublicPosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	// Get public posts(paginated)
	var publicPosts = []models.Post{}
	if err := configs.Database.Model(&models.Post{}).Preload("Media").Offset(offset).Limit(limit).Order("created_at DESC").Find(&publicPosts, "profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of public posts
	var numPublicPosts int64
	if err := configs.Database.Model(&models.Post{}).Where("profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, false).Count(&numPublicPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numPublicPosts) / float64(limit))),
			"data":         publicPosts,
		},
	}))
}

func GetExclusivePosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	// Get exclusive posts(paginated)
	var exclusivePosts = []models.Post{}
	if err := configs.Database.Model(&models.Post{}).Preload("Media").Offset(offset).Limit(limit).Order("created_at DESC").Find(&exclusivePosts, "profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, true).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of exclusive posts
	var numExclusivePosts int64
	if err := configs.Database.Model(&models.Post{}).Where("profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, true).Count(&numExclusivePosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numExclusivePosts) / float64(limit))),
			"data":         exclusivePosts,
		},
	}))
}
