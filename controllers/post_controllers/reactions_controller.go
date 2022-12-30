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
	var dislikeObj models.PostDislike
	if err := configs.Database.Model(&models.PostDislike{}).Delete(&dislikeObj, "post_id = ? AND disliker_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	newLikeObj := models.PostLike{
		PostId:    c.Params("postId"),
		LikerId:   reqProfile.Id,
		CreatedAt: time.Now(),
	}
	if err := configs.Database.Model(&models.PostLike{}).Where("post_id = ? AND liker_id = ?", newLikeObj.PostId, newLikeObj.LikerId).FirstOrCreate(&newLikeObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been liked."}))
}

func DislikePost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the like object(if it exists)
	var likeObj models.PostLike
	if err := configs.Database.Model(&models.PostLike{}).Delete(&likeObj, "post_id = ? AND liker_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	newDislikeObj := models.PostDislike{
		PostId:     c.Params("postId"),
		DislikerId: reqProfile.Id,
		CreatedAt:  time.Now(),
	}
	if err := configs.Database.Model(&models.PostDislike{}).Where("post_id = ? AND disliker_id = ?", newDislikeObj.PostId, newDislikeObj.DislikerId).FirstOrCreate(&newDislikeObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been disliked."}))
}

func RemoveLike(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	var likeObj models.PostLike
	if err := configs.Database.Model(&models.PostLike{}).Delete(&likeObj, "post_id = ? AND liker_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Like has been removed."}))
}

func RemoveDislike(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	var dislikeObj models.PostDislike
	if err := configs.Database.Model(&models.PostDislike{}).Delete(&dislikeObj, "post_id = ? AND disliker_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
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
			"ON post_likes.post_id = \"%s\" AND profile.id = post_likes.liker_id "+
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
			"ON post_dislikes.post_id = \"%s\" AND profile.id = post_dislikes.disliker_id "+
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
	// var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Get Liked Posts(paginated)
	var likedPosts = []models.Post{}

	// Get total number of liked posts
	var numLikedPosts int64
	if err := configs.Database.Model(&models.PostLike{}).Where("liker_id = ?", reqProfile.Id).Count(&numLikedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numLikedPosts) / float64(limit))),
			"data":         likedPosts,
		},
	}))
}

func GetDisikedPosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	// var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Get Liked Posts(paginated)
	var dislikedPosts = []models.Post{}

	// Get total number of liked posts
	var numDislikedPosts int64
	if err := configs.Database.Model(&models.PostDislike{}).Where("disliker_id = ?", reqProfile.Id).Count(&numDislikedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numDislikedPosts) / float64(limit))),
			"data":         dislikedPosts,
		},
	}))
}
