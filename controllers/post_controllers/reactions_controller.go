package postcontrollers

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func LikePost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the dislike object(if it exists)
	var dislikeObj models.PostDislike
	if err := configs.Database.Table("post_dislikes").Delete(&dislikeObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	newLikeObj := models.PostLike{
		PostId:    c.Params("postId"),
		ProfileId: reqProfile.Id,
		CreatedAt: time.Now(),
	}
	if err := configs.Database.Table("post_likes").Where("post_id = ? AND profile_id = ?", newLikeObj.PostId, newLikeObj.ProfileId).FirstOrCreate(&newLikeObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been liked."}))
}

func DislikePost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the like object(if it exists)
	var likeObj models.PostLike
	if err := configs.Database.Table("post_likes").Delete(&likeObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	newDislikeObj := models.PostDislike{
		PostId:    c.Params("postId"),
		ProfileId: reqProfile.Id,
		CreatedAt: time.Now(),
	}
	if err := configs.Database.Table("post_dislikes").Where("post_id = ? AND profile_id = ?", newDislikeObj.PostId, newDislikeObj.ProfileId).FirstOrCreate(&newDislikeObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been disliked."}))
}

func RemoveLike(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	var likeObj models.PostLike
	if err := configs.Database.Table("post_likes").Delete(&likeObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Like has been removed."}))
}

func RemoveDislike(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	var dislikeObj models.PostDislike
	if err := configs.Database.Table("post_dislikes").Delete(&dislikeObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Dislike has been removed."}))
}

func GetLikesOfPost(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	// Get Liked Posts(paginated)
	var postLikers = []models.MiniProfile{}
	query := fmt.Sprintf(
		"SELECT profile.id, profile.username, profile.name, profile.mini_avatar "+
			"FROM post_likes "+
			"JOIN profiles as profile "+
			"ON post_likes.post_id = \"%s\" AND profile.id = post_likes.profile_id "+
			"ORDER BY post_likes.created_at DESC LIMIT %d OFFSET %d",
		c.Params("postId"), limit, offset,
	)
	if err := configs.Database.Raw(query).Scan(&postLikers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of likes
	var post = models.Post{Base: models.Base{Id: c.Params("postId")}}
	numLikes := configs.Database.Model(&post).Association("Likes").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numLikes) / float64(limit))),
			"data":         postLikers,
		},
	}))
}

func GetDislikesOfPost(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	// Get Liked Posts(paginated)
	var postDislikers = []models.MiniProfile{}
	query := fmt.Sprintf(
		"SELECT profile.id, profile.username, profile.name, profile.mini_avatar "+
			"FROM post_dislikes "+
			"JOIN profiles as profile "+
			"ON post_dislikes.post_id = \"%s\" AND profile.id = post_dislikes.profile_id "+
			"ORDER BY post_dislikes.created_at DESC LIMIT %d OFFSET %d",
		c.Params("postId"), limit, offset,
	)
	if err := configs.Database.Raw(query).Scan(&postDislikers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of likes
	var post = models.Post{Base: models.Base{Id: c.Params("postId")}}
	numDislikes := configs.Database.Model(&post).Association("Dislikes").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
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
			"(SELECT GROUP_CONCAT(m.media_url, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS media_urls, "+
			"(SELECT GROUP_CONCAT(m.is_image, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS is_images, "+
			"(SELECT GROUP_CONCAT(m.is_video, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS is_videos, "+
			"(SELECT GROUP_CONCAT(m.is_audio, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS is_audios, "+
			"COUNT(l.post_id) AS num_likes, COUNT(d.post_id) AS num_dislikes, COUNT(b.post_id) AS num_bookmarks, "+
			"SUM(CASE WHEN d.profile_id = \"%s\" THEN 1 ELSE 0 END) AS is_disliked, "+
			"SUM(CASE WHEN b.profile_id = \"%s\" THEN 1 ELSE 0 END) AS is_bookmarked "+
			"FROM post_likes l "+
			"JOIN posts p ON l.post_id = p.id "+
			"JOIN profiles u ON p.profile_id = u.id "+
			"LEFT JOIN post_bookmarks b ON p.id = b.post_id "+
			"LEFT JOIN post_dislikes d ON p.id = d.post_id "+
			"WHERE l.profile_id = \"%s\" "+
			"GROUP BY p.id, u.id "+
			"ORDER BY l.created_at DESC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, reqProfile.Id, limit, offset,
	)
	var unpreparedLikedPosts = []postsWithoutMedia{}
	if err := configs.Database.Raw(query).Scan(&unpreparedLikedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	var likedPosts = []postsWithMedia{}
	for _, post := range unpreparedLikedPosts { // The limit is capped at 25 and the post_media for a post is capped at 5. This loop has 125 iterations at most
		post.IsLiked = true
		var mediaObjs = []miniPostMedia{}
		mediaUrls := strings.Split(post.MediaUrls, ",")
		isImages := strings.Split(post.IsImages, ",")
		isVideos := strings.Split(post.IsVideos, ",")
		isAudios := strings.Split(post.IsAudios, ",")
		for i, url := range mediaUrls {
			mediaObjs = append(mediaObjs, miniPostMedia{MediaUrl: url, IsImage: stringToBool(isImages[i]), IsVideo: stringToBool(isVideos[i]), IsAudio: stringToBool(isAudios[i])})
		}
		likedPosts = append(likedPosts, postsWithMedia{postsWithoutMedia: post, Media: mediaObjs})
	}

	var numLikes int64
	if err := configs.Database.Table("post_likes").Where("profile_id = ?", reqProfile.Id).Count(&numLikes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
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
			"(SELECT GROUP_CONCAT(m.media_url, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS media_urls, "+
			"(SELECT GROUP_CONCAT(m.is_image, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS is_images, "+
			"(SELECT GROUP_CONCAT(m.is_video, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS is_videos, "+
			"(SELECT GROUP_CONCAT(m.is_audio, '') FROM post_media m WHERE m.post_id = p.id ORDER BY m.position) AS is_audios, "+
			"COUNT(l.post_id) AS num_likes, COUNT(d.post_id) AS num_dislikes, COUNT(b.post_id) AS num_bookmarks, "+
			"SUM(CASE WHEN l.profile_id = \"%s\" THEN 1 ELSE 0 END) AS is_liked, "+
			"SUM(CASE WHEN b.profile_id = \"%s\" THEN 1 ELSE 0 END) AS is_bookmarked "+
			"FROM post_dislikes d "+
			"JOIN posts p ON d.post_id = p.id "+
			"JOIN profiles u ON p.profile_id = u.id "+
			"LEFT JOIN post_likes l ON p.id = l.post_id "+
			"LEFT JOIN post_bookmarks b ON p.id = b.post_id "+
			"WHERE d.profile_id = \"%s\" "+
			"GROUP BY p.id, u.id "+
			"ORDER BY d.created_at DESC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, reqProfile.Id, limit, offset,
	)
	var unpreparedDislikedPosts = []postsWithoutMedia{}
	if err := configs.Database.Raw(query).Scan(&unpreparedDislikedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	var dislikedPosts = []postsWithMedia{}
	for _, post := range unpreparedDislikedPosts { // The limit is capped at 25 and the post_media for a post is capped at 5. This loop has 125 iterations at most
		post.IsDisliked = true
		var mediaObjs = []miniPostMedia{}
		mediaUrls := strings.Split(post.MediaUrls, ",")
		isImages := strings.Split(post.IsImages, ",")
		isVideos := strings.Split(post.IsVideos, ",")
		isAudios := strings.Split(post.IsAudios, ",")
		for i, url := range mediaUrls {
			mediaObjs = append(mediaObjs, miniPostMedia{MediaUrl: url, IsImage: stringToBool(isImages[i]), IsVideo: stringToBool(isVideos[i]), IsAudio: stringToBool(isAudios[i])})
		}
		dislikedPosts = append(dislikedPosts, postsWithMedia{postsWithoutMedia: post, Media: mediaObjs})
	}

	var numDislikes int64
	if err := configs.Database.Table("post_dislikes").Where("profile_id = ?", reqProfile.Id).Count(&numDislikes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numDislikes) / float64(limit))),
			"data":         dislikedPosts,
		},
	}))
}
