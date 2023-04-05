package postcontrollers

import (
	"math"
	"strings"
	"sync"
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
	query := "DELETE FROM comments USING posts WHERE comments.id = ? AND posts.id = comments.post_id AND posts.profile_id = ?"
	if err := configs.Database.WithContext(dbCtx).Exec(query, c.Params("commentId"), reqProfile.Id).Error; err != nil {
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

func GetComments(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Run both queries concurrently to reduce response time
	errChan := make(chan error, 1) // make this buffered so that the goroutine doesn't block
	wg := new(sync.WaitGroup)
	wg.Add(1)
	var comments = []responses.Comment{}
	go func() {
		defer wg.Done()

		query := "SELECT c.id AS comment_id, c.body AS body, c.created_at AS created_at, c.is_edited AS is_edited, "
		query += "p.id AS profile_id, p.username AS profile_username, p.mini_avatar AS profile_mini_avatar, "
		query += "COALESCE(cl.likes, 0) as num_likes, COALESCE(cd.dislikes, 0) as num_dislikes, COALESCE(cr.replies, 0) as num_replies, "
		query += "CASE WHEN dl.profile_id IS NOT NULL THEN true ELSE false END AS is_liked, "
		query += "CASE WHEN dd.profile_id IS NOT NULL THEN true ELSE false END AS is_disliked "

		query += "FROM comments c "
		query += "JOIN profiles p ON c.commenter_id = p.id "
		query += "LEFT JOIN (SELECT comment_id, COUNT(*) as likes FROM comment_likes GROUP BY comment_id) cl ON c.id = cl.comment_id "
		query += "LEFT JOIN (SELECT comment_id, COUNT(*) as dislikes FROM comment_dislikes GROUP BY comment_id) cd ON c.id = cd.comment_id "
		query += "LEFT JOIN (SELECT comment_replied_to_id, COUNT(*) as replies FROM comments WHERE comment_replied_to_id IS NOT NULL GROUP BY comment_replied_to_id) cr ON c.id = cr.comment_replied_to_id "
		query += "LEFT JOIN comment_likes dl ON c.id = dl.comment_id AND dl.profile_id = ? "
		query += "LEFT JOIN comment_dislikes dd ON c.id = dd.comment_id AND dd.profile_id = ? "

		query += "WHERE c.post_id = ? AND c.comment_replied_to_id IS NULL "
		query += "ORDER BY c.created_at DESC "
		query += "LIMIT ? OFFSET ?;"

		dbCtx, dbCancel := configs.NewQueryContext()
		defer dbCancel()
		errChan <- configs.Database.WithContext(dbCtx).Raw(query, reqProfile.Id, reqProfile.Id, c.Params("postId"), limit, offset).Scan(&comments).Error
	}()

	// Get total number of comments exlcuding replies
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numComments int64
	if err := configs.Database.WithContext(dbCtx2).Table("comments").Where("post_id = ? AND comment_replied_to_id IS NULL", c.Params("postId")).Count(&numComments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	wg.Wait()

	if err := <-errChan; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	close(errChan)

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

	// Run both queries concurrently to reduce response time
	errChan := make(chan error, 1) // make this buffered so that the goroutine doesn't block
	wg := new(sync.WaitGroup)
	wg.Add(1)
	var replies = []responses.Comment{}
	go func() {
		defer wg.Done()

		query := "SELECT c.id AS comment_id, c.body AS body, c.created_at AS created_at, c.is_edited AS is_edited, "
		query += "p.id AS profile_id, p.username AS profile_username, p.mini_avatar AS profile_mini_avatar, "
		query += "COALESCE(cl.likes, 0) as num_likes, COALESCE(cd.dislikes, 0) as num_dislikes, COALESCE(cr.replies, 0) as num_replies, "
		query += "CASE WHEN dl.profile_id IS NOT NULL THEN true ELSE false END AS is_liked, "
		query += "CASE WHEN dd.profile_id IS NOT NULL THEN true ELSE false END AS is_disliked "

		query += "FROM comments c "
		query += "JOIN profiles p ON c.commenter_id = p.id "
		query += "LEFT JOIN (SELECT comment_id, COUNT(*) as likes FROM comment_likes GROUP BY comment_id) cl ON c.id = cl.comment_id "
		query += "LEFT JOIN (SELECT comment_id, COUNT(*) as dislikes FROM comment_dislikes GROUP BY comment_id) cd ON c.id = cd.comment_id "
		query += "LEFT JOIN (SELECT comment_replied_to_id, COUNT(*) as replies FROM comments WHERE comment_replied_to_id IS NOT NULL GROUP BY comment_replied_to_id) cr ON c.id = cr.comment_replied_to_id "
		query += "LEFT JOIN comment_likes dl ON c.id = dl.comment_id AND dl.profile_id = ? "
		query += "LEFT JOIN comment_dislikes dd ON c.id = dd.comment_id AND dd.profile_id = ? "

		query += "WHERE c.comment_replied_to_id = ? "
		query += "ORDER BY c.created_at ASC "
		query += "LIMIT ? OFFSET ?;"

		dbCtx, dbCancel := configs.NewQueryContext()
		defer dbCancel()
		errChan <- configs.Database.WithContext(dbCtx).Raw(query, reqProfile.Id, reqProfile.Id, c.Params("commentId"), limit, offset).Scan(&replies).Error
	}()

	// Get total number of replies
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numReplies int64
	if err := configs.Database.WithContext(dbCtx2).Table("comments").Where("comment_replied_to_id = ?", c.Params("commentId")).Count(&numReplies).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	wg.Wait()

	if err := <-errChan; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	close(errChan)

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numReplies) / float64(limit))),
			"data":         replies,
		},
	}))
}
