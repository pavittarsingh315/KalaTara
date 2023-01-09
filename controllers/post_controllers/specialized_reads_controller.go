package postcontrollers

import (
	"fmt"
	"math"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func GetFollowingsFeed(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get followings feed."}))
}

func GetSubscriptionsFeed(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get subscriptions feed."}))
}

func GetArchivedPosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	query := fmt.Sprintf(
		"SELECT p.id AS post_id, p.title AS post_title, p.caption AS post_caption, p.created_at AS created_at, "+
			"u.id AS profile_id, u.username AS profile_username, u.name AS profile_name, u.mini_avatar AS profile_mini_avatar, "+
			"(SELECT GROUP_CONCAT(m.media_url, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS media_urls, "+
			"(SELECT GROUP_CONCAT(m.is_image, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS is_images, "+
			"(SELECT GROUP_CONCAT(m.is_video, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS is_videos, "+
			"(SELECT GROUP_CONCAT(m.is_audio, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS is_audios, "+
			"l.num_likes, d.num_dislikes, b.num_bookmarks, "+
			"(CASE WHEN (SELECT profile_id FROM post_likes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_liked, "+
			"(CASE WHEN (SELECT profile_id FROM post_dislikes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_disliked, "+
			"(CASE WHEN (SELECT profile_id FROM post_bookmarks WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_bookmarked "+
			"FROM posts p "+
			"JOIN profiles u ON p.profile_id = u.id AND u.id = \"%s\" AND p.is_archived = %d "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_likes FROM post_likes GROUP by post_id) l ON p.id = l.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_dislikes FROM post_dislikes GROUP by post_id) d ON p.id = d.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_bookmarks FROM post_bookmarks GROUP by post_id) b ON p.id = b.post_id "+
			"GROUP BY p.id, u.id "+
			"ORDER BY p.created_at DESC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, reqProfile.Id, reqProfile.Id, 1, limit, offset,
	)
	var unpreparedArchivedPosts = []postsWithoutMedia{}
	if err := configs.Database.Raw(query).Scan(&unpreparedArchivedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	var archivedPosts = preparePosts(&unpreparedArchivedPosts, false, false, false)

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
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get public posts."}))
}

func GetExclusivePosts(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get exclusive posts."}))
}
