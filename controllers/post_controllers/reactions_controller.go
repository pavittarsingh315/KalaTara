package postcontrollers

import (
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func LikePost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the dislike object(if it exists)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var dislikeObj models.PostDislike
	if err := configs.Database.WithContext(dbCtx).Table("post_dislikes").Delete(&dislikeObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	newLikeObj := models.PostLike{
		PostId:    c.Params("postId"),
		ProfileId: reqProfile.Id,
		CreatedAt: time.Now(),
	}
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	if err := configs.Database.WithContext(dbCtx2).Table("post_likes").Where("post_id = ? AND profile_id = ?", newLikeObj.PostId, newLikeObj.ProfileId).FirstOrCreate(&newLikeObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been liked."}))
}

func DislikePost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the like object(if it exists)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var likeObj models.PostLike
	if err := configs.Database.WithContext(dbCtx).Table("post_likes").Delete(&likeObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	newDislikeObj := models.PostDislike{
		PostId:    c.Params("postId"),
		ProfileId: reqProfile.Id,
		CreatedAt: time.Now(),
	}
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	if err := configs.Database.WithContext(dbCtx2).Table("post_dislikes").Where("post_id = ? AND profile_id = ?", newDislikeObj.PostId, newDislikeObj.ProfileId).FirstOrCreate(&newDislikeObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been disliked."}))
}

func RemoveLike(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var likeObj models.PostLike
	if err := configs.Database.WithContext(dbCtx).Table("post_likes").Delete(&likeObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Like has been removed."}))
}

func RemoveDislike(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var dislikeObj models.PostDislike
	if err := configs.Database.WithContext(dbCtx).Table("post_dislikes").Delete(&dislikeObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Dislike has been removed."}))
}

func GetLikesOfPost(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	query := configs.Database.Table("post_likes").
		Select("profile.id, profile.username, profile.name, profile.mini_avatar").
		Joins("JOIN profiles as profile ON post_likes.post_id = ? AND profile.id = post_likes.profile_id", c.Params("postId")).
		Order("post_likes.created_at DESC").
		Limit(limit).Offset(offset)

	// Get likers(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var postLikers = []models.MiniProfile{}
	if err := query.WithContext(dbCtx).Scan(&postLikers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of likes
	var post = models.Post{Base: models.Base{Id: c.Params("postId")}}
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	numLikes := configs.Database.WithContext(dbCtx2).Model(&post).Association("Likes").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numLikes) / float64(limit))),
			"data":         postLikers,
		},
	}))
}

func GetDislikesOfPost(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	query := configs.Database.Table("post_dislikes").
		Select("profile.id, profile.username, profile.name, profile.mini_avatar").
		Joins("JOIN profiles as profile ON post_dislikes.post_id = ? AND profile.id = post_dislikes.profile_id", c.Params("postId")).
		Order("post_dislikes.created_at DESC").
		Limit(limit).Offset(offset)

	// Get dislikers(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var postDislikers = []models.MiniProfile{}
	if err := query.WithContext(dbCtx).Scan(&postDislikers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of dislikes
	var post = models.Post{Base: models.Base{Id: c.Params("postId")}}
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	numDislikes := configs.Database.WithContext(dbCtx2).Model(&post).Association("Dislikes").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numDislikes) / float64(limit))),
			"data":         postDislikers,
		},
	}))
}

func GetLikedPosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	query := fmt.Sprintf(
		"SELECT p.id AS post_id, p.title AS post_title, p.caption AS post_caption, p.created_at AS created_at, "+
			"u.id AS profile_id, u.username AS profile_username, u.name AS profile_name, u.mini_avatar AS profile_mini_avatar, "+
			"(SELECT GROUP_CONCAT(CONCAT(m.media_url, '') ORDER BY m.position) FROM post_media m WHERE m.post_id = p.id) AS media_urls, "+
			"(SELECT GROUP_CONCAT(CONCAT(m.is_image, '') ORDER BY m.position) FROM post_media m WHERE m.post_id = p.id) AS is_images, "+
			"(SELECT GROUP_CONCAT(CONCAT(m.is_video, '') ORDER BY m.position) FROM post_media m WHERE m.post_id = p.id) AS is_videos, "+
			"(SELECT GROUP_CONCAT(CONCAT(m.is_audio, '') ORDER BY m.position) FROM post_media m WHERE m.post_id = p.id) AS is_audios, "+
			"l2.num_likes, d.num_dislikes, b.num_bookmarks, "+
			"(CASE WHEN (SELECT profile_id FROM post_dislikes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_disliked, "+
			"(CASE WHEN (SELECT profile_id FROM post_bookmarks WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_bookmarked "+
			"FROM post_likes l "+
			"JOIN posts p ON l.post_id = p.id AND l.profile_id = \"%s\" AND p.is_archived = %d "+
			"JOIN profiles u ON p.profile_id = u.id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_likes FROM post_likes GROUP by post_id) l2 ON p.id = l2.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_dislikes FROM post_dislikes GROUP by post_id) d ON p.id = d.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_bookmarks FROM post_bookmarks GROUP by post_id) b ON p.id = b.post_id "+
			"GROUP BY p.id, u.id "+
			"ORDER BY l.created_at DESC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, reqProfile.Id, 0, limit, offset,
	)
	var unpreparedLikedPosts = []postsWithoutMedia{}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Raw(query).Scan(&unpreparedLikedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	var likedPosts = preparePosts(&unpreparedLikedPosts, false, true, false)

	var numLikes int64
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	if err := configs.Database.WithContext(dbCtx2).Table("post_likes").Where("profile_id = ?", reqProfile.Id).Count(&numLikes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numLikes) / float64(limit))),
			"data":         likedPosts,
		},
	}))
}

func GetDisikedPosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	query := fmt.Sprintf(
		"SELECT p.id AS post_id, p.title AS post_title, p.caption AS post_caption, p.created_at AS created_at, "+
			"u.id AS profile_id, u.username AS profile_username, u.name AS profile_name, u.mini_avatar AS profile_mini_avatar, "+
			"(SELECT GROUP_CONCAT(CONCAT(m.media_url, '') ORDER BY m.position) FROM post_media m WHERE m.post_id = p.id) AS media_urls, "+
			"(SELECT GROUP_CONCAT(CONCAT(m.is_image, '') ORDER BY m.position) FROM post_media m WHERE m.post_id = p.id) AS is_images, "+
			"(SELECT GROUP_CONCAT(CONCAT(m.is_video, '') ORDER BY m.position) FROM post_media m WHERE m.post_id = p.id) AS is_videos, "+
			"(SELECT GROUP_CONCAT(CONCAT(m.is_audio, '') ORDER BY m.position) FROM post_media m WHERE m.post_id = p.id) AS is_audios, "+
			"l.num_likes, d2.num_dislikes, b.num_bookmarks, "+
			"(CASE WHEN (SELECT profile_id FROM post_likes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_liked, "+
			"(CASE WHEN (SELECT profile_id FROM post_bookmarks WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_bookmarked "+
			"FROM post_dislikes d "+
			"JOIN posts p ON d.post_id = p.id AND d.profile_id = \"%s\" AND p.is_archived = %d"+
			"JOIN profiles u ON p.profile_id = u.id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_likes FROM post_likes GROUP by post_id) l ON p.id = l.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_dislikes FROM post_dislikes GROUP by post_id) d2 ON p.id = d2.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_bookmarks FROM post_bookmarks GROUP by post_id) b ON p.id = b.post_id "+
			"GROUP BY p.id, u.id "+
			"ORDER BY d.created_at DESC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, reqProfile.Id, 0, limit, offset,
	)
	var unpreparedDislikedPosts = []postsWithoutMedia{}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Raw(query).Scan(&unpreparedDislikedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	var dislikedPosts = preparePosts(&unpreparedDislikedPosts, false, false, true)

	var numDislikes int64
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	if err := configs.Database.WithContext(dbCtx2).Table("post_dislikes").Where("profile_id = ?", reqProfile.Id).Count(&numDislikes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numDislikes) / float64(limit))),
			"data":         dislikedPosts,
		},
	}))
}
