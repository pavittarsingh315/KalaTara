package postcontrollers

import (
	"fmt"
	"math"
	"strings"
	"time"

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
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}, err))
	}

	if reqBody.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}, nil))
	}

	reqBody.Body = strings.TrimSpace(reqBody.Body) // remove leading and trailing whitespace

	length := uniseg.GraphemeClusterCount(reqBody.Body)
	if length > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Comment is too long."}, nil))
	}

	newComment := models.Comment{
		PostId:             c.Params("postId"),
		CommenterId:        reqProfile.Id,
		CommentRepliedToId: reqBody.RepliesTo,
		Body:               reqBody.Body,
		IsEdited:           false,
	}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Model(&models.Comment{}).Create(&newComment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": newComment}))
}

func EditComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)
	reqBody := struct {
		Body string `json:"body"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}, err))
	}

	if reqBody.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}, nil))
	}

	reqBody.Body = strings.TrimSpace(reqBody.Body) // remove leading and trailing whitespace

	length := uniseg.GraphemeClusterCount(reqBody.Body)
	if length > 500 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Comment is too long."}, nil))
	}

	// Update the fields
	var base = models.Base{Id: c.Params("commentId")}
	var comment = models.Comment{Base: base, CommenterId: reqProfile.Id}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Model(&comment).Updates(map[string]interface{}{"is_edited": true, "body": reqBody.Body}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Comment has been successfully updated."}))
}

// Commenter deletes their own comment
func DeleteComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var comment models.Comment
	if err := configs.Database.WithContext(dbCtx).Model(&models.Comment{}).Delete(&comment, "id = ? AND commenter_id = ?", c.Params("commentId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Comment Deleted."}))
}

// Post owner deletes a comment
func RemoveComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	query := fmt.Sprintf("DELETE c FROM comments c JOIN posts p ON c.id = \"%s\" AND c.post_id = p.id AND p.profile_id = \"%s\"", c.Params("commentId"), reqProfile.Id)
	if err := configs.Database.WithContext(dbCtx).Exec(query).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Comment Removed."}))
}

func LikeComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the dislike object(if it exists)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var dislikeObj models.CommentDislike
	if err := configs.Database.WithContext(dbCtx).Table("comment_dislikes").Delete(&dislikeObj, "comment_id = ? AND profile_id = ?", c.Params("commentId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	newLikeObj := models.CommentLike{
		CommentId: c.Params("commentId"),
		ProfileId: reqProfile.Id,
		CreatedAt: time.Now(),
	}
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	if err := configs.Database.WithContext(dbCtx2).Table("comment_likes").Where("comment_id = ? AND profile_id = ?", newLikeObj.CommentId, newLikeObj.ProfileId).FirstOrCreate(&newLikeObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Comment has been liked."}))
}

func DislikeComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the like object(if it exists)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var likeObj models.CommentLike
	if err := configs.Database.WithContext(dbCtx).Table("comment_likes").Delete(&likeObj, "comment_id = ? AND profile_id = ?", c.Params("commentId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	newDislikeObj := models.CommentDislike{
		CommentId: c.Params("commentId"),
		ProfileId: reqProfile.Id,
		CreatedAt: time.Now(),
	}
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	if err := configs.Database.WithContext(dbCtx2).Table("comment_dislikes").Where("comment_id = ? AND profile_id = ?", newDislikeObj.CommentId, newDislikeObj.ProfileId).FirstOrCreate(&newDislikeObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Comment has been disliked."}))
}

func RemoveLikeFromComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the like object(if it exists)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var likeObj models.CommentLike
	if err := configs.Database.WithContext(dbCtx).Table("comment_likes").Delete(&likeObj, "comment_id = ? AND profile_id = ?", c.Params("commentId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Like has been removed."}))
}

func RemoveDislikeFromComment(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the dislike object(if it exists)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var dislikeObj models.CommentDislike
	if err := configs.Database.WithContext(dbCtx).Table("comment_dislikes").Delete(&dislikeObj, "comment_id = ? AND profile_id = ?", c.Params("commentId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Dislike has been removed."}))
}

type commentResponseObject struct {
	CommentId string    `json:"comment_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	IsEdited  bool      `json:"is_edited"`

	ProfileId  string `json:"profile_id"`
	Username   string `json:"username" gorm:"column:profile_username"`
	MiniAvatar string `json:"mini_avatar" gorm:"column:profile_mini_avatar"`

	NumLikes    int `json:"num_likes"`
	NumDislikes int `json:"num_dislikes"`
	NumReplies  int `json:"num_replies"`

	IsLiked    bool `json:"is_liked"`
	IsDisliked bool `json:"is_disliked"`
}

func GetComments(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	query := fmt.Sprintf(
		"SELECT c.id AS comment_id, c.body AS body, c.created_at AS created_at, c.is_edited AS is_edited, "+
			"p.id AS profile_id, p.username AS profile_username, p.mini_avatar AS profile_mini_avatar, "+
			"l.num_likes, d.num_dislikes, r.num_replies, "+
			"(CASE WHEN (SELECT profile_id FROM comment_likes WHERE comment_id = c.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_liked, "+
			"(CASE WHEN (SELECT profile_id FROM comment_dislikes WHERE comment_id = c.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_disliked "+
			"FROM comments c "+
			"JOIN profiles p ON p.id = c.commenter_id AND c.comment_replied_to_id IS NULL AND c.post_id = \"%s\" "+
			"LEFT JOIN (SELECT comment_id, COUNT(*) AS num_likes FROM comment_likes GROUP by comment_id) l ON c.id = l.comment_id "+
			"LEFT JOIN (SELECT comment_id, COUNT(*) AS num_dislikes FROM comment_dislikes GROUP by comment_id) d ON c.id = d.comment_id "+
			"LEFT JOIN (SELECT comment_replied_to_id, COUNT(*) AS num_replies FROM comments GROUP by comment_replied_to_id) r ON c.id = r.comment_replied_to_id "+
			"ORDER BY c.created_at DESC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, c.Params("postId"), limit, offset,
	)
	var comments = []commentResponseObject{}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Raw(query).Scan(&comments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of likes
	var post = models.Post{Base: models.Base{Id: c.Params("postId")}}
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	numComments := configs.Database.WithContext(dbCtx2).Model(&post).Association("Comments").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numComments) / float64(limit))),
			"data":         comments,
		},
	}))
}

func GetReplies(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	query := fmt.Sprintf(
		"SELECT c.id AS comment_id, c.body AS body, c.created_at AS created_at, c.is_edited AS is_edited, "+
			"p.id AS profile_id, p.username AS profile_username, p.mini_avatar AS profile_mini_avatar, "+
			"l.num_likes, d.num_dislikes, r.num_replies, "+
			"(CASE WHEN (SELECT profile_id FROM comment_likes WHERE comment_id = c.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_liked, "+
			"(CASE WHEN (SELECT profile_id FROM comment_dislikes WHERE comment_id = c.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_disliked "+
			"FROM comments c "+
			"JOIN profiles p ON p.id = c.commenter_id AND c.comment_replied_to_id = \"%s\" "+
			"LEFT JOIN (SELECT comment_id, COUNT(*) AS num_likes FROM comment_likes GROUP by comment_id) l ON c.id = l.comment_id "+
			"LEFT JOIN (SELECT comment_id, COUNT(*) AS num_dislikes FROM comment_dislikes GROUP by comment_id) d ON c.id = d.comment_id "+
			"LEFT JOIN (SELECT comment_replied_to_id, COUNT(*) AS num_replies FROM comments GROUP by comment_replied_to_id) r ON c.id = r.comment_replied_to_id "+
			"ORDER BY c.created_at ASC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, c.Params("commentId"), limit, offset,
	)
	var replies = []commentResponseObject{}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Raw(query).Scan(&replies).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of replies
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numReplies int64
	if err := configs.Database.WithContext(dbCtx2).Model(&models.Comment{}).Where("comment_replied_to_id = ?", c.Params("commentId")).Count(&numReplies).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numReplies) / float64(limit))),
			"data":         replies,
		},
	}))
}
