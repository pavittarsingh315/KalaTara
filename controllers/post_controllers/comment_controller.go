package postcontrollers

import (
	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/responses"
)

func CreateComment(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Create Comment."}))
}

func DeleteComment(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Delete Comment."}))
}

func EditComment(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Edit Comment."}))
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
