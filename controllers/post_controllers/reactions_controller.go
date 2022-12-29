package postcontrollers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

// IF EXISTS(SELECT 1 FROM posts as post WHERE post.id = \"%s\")
// AND NOT EXISTS(SELECT 1 FROM post_likes as pl WHERE pl.post_id = \"%s\" AND pl.liker_id = \"%s\")
// BEGIN
//     INSERT INTO post_likes (post_id, liker_id, created_at) VALUES (\"%s\", \"%s\", \"%s\")
//     DELETE FROM post_dislikes WHERE post_id = \"%s\" AND disliker_id = \"%s\"
// END

/*
	SELECT IF(
      EXISTS(SELECT 1 FROM posts as post WHERE post.id = "b747fc31-58bb-4a32-9f28-7029d39013d0")
      AND NOT EXISTS(SELECT 1 FROM post_likes as pl WHERE pl.post_id = "b747fc31-58bb-4a32-9f28-7029d39013d0" AND pl.liker_id = "b9613d83-8fc1-4c4f-ba65-175ebe8dc0ba")
      , 1, 0
   );
*/

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
