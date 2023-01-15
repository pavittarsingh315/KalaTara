package postcontrollers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rivo/uniseg"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func CreateComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)
	reqBody := struct {
		Body      string  `json:"body"`
		RepliesTo *string `json:"replies_to"` // this is allowed to be nil
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	if reqBody.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}

	reqBody.Body = strings.TrimSpace(reqBody.Body) // remove leading and trailing whitespace

	length := uniseg.GraphemeClusterCount(reqBody.Body)
	if length > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Comment is too long."}))
	}

	newComment := models.Comment{
		PostId:             c.Params("postId"),
		CommenterId:        reqProfile.Id,
		CommentRepliedToId: reqBody.RepliesTo,
		Body:               reqBody.Body,
		IsEdited:           false,
	}
	if err := configs.Database.Model(&models.Comment{}).Create(&newComment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": newComment}))
}

func EditComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)
	reqBody := struct {
		Body string `json:"body"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	if reqBody.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}

	reqBody.Body = strings.TrimSpace(reqBody.Body) // remove leading and trailing whitespace

	length := uniseg.GraphemeClusterCount(reqBody.Body)
	if length > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Comment is too long."}))
	}

	// Update the fields
	var base = models.Base{Id: c.Params("commentId")}
	var comment = models.Comment{Base: base, CommenterId: reqProfile.Id}
	if err := configs.Database.Model(&comment).Updates(map[string]interface{}{"is_edited": true, "body": reqBody.Body}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Comment has been successfully updated."}))
}

// Commenter deletes their own comment
func DeleteComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var comment models.Comment
	if err := configs.Database.Model(&models.Comment{}).Delete(&comment, "id = ? AND commenter_id = ?", c.Params("commentId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Comment Deleted."}))
}

// Post owner deletes a comment
func RemoveComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	query := fmt.Sprintf("DELETE c FROM comments c JOIN posts p ON c.id = \"%s\" AND c.post_id = p.id AND p.profile_id = \"%s\"", c.Params("commentId"), reqProfile.Id)
	if err := configs.Database.Exec(query).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Comment Removed."}))
}

func LikeComment(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Like Comment."}))
}

func DislikeComment(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Dislike Comment."}))
}

func RemoveLikeFromComment(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Remove Like From Comment."}))
}

func RemoveDislikeFromComment(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Remove Dislike From Comment."}))
}

func GetComments(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get Comments."}))
}

func GetReplies(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get Replies."}))
}
