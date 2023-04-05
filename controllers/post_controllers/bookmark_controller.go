package postcontrollers

import (
	"math"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func BookmarkPost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	newBookmarkObj := models.PostBookmark{
		PostId:    c.Params("postId"),
		ProfileId: reqProfile.Id,
		CreatedAt: time.Now(),
	}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Table("post_bookmarks").Where("post_id = ? AND profile_id = ?", newBookmarkObj.PostId, newBookmarkObj.ProfileId).FirstOrCreate(&newBookmarkObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been bookmarked."}))
}

func RemoveBookmark(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var bookmarkObj models.PostBookmark
	if err := configs.Database.WithContext(dbCtx).Table("post_bookmarks").Delete(&bookmarkObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Bookmark has been removed."}))
}

func GetBookmarkedPosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Run both queries concurrently to reduce response time
	errChan := make(chan error, 1) // make this buffered so that the goroutine doesn't block
	wg := new(sync.WaitGroup)
	wg.Add(1)
	var bookmarkedPosts = []responses.Post{}
	go func() {
		defer wg.Done()

		query := "WITH media_agg AS (SELECT post_id, json_agg(json_build_object('media_url', media_url, 'is_image', is_image, 'is_video', is_video, 'is_audio', is_audio) ORDER BY position) AS media_data FROM post_media GROUP BY post_id), "
		query += "likes_agg AS (SELECT post_id, COUNT(*) AS likes FROM post_likes GROUP BY post_id), "
		query += "dislikes_agg AS (SELECT post_id, COUNT(*) AS dislikes FROM post_dislikes GROUP BY post_id), "
		query += "bookmarks_agg AS (SELECT post_id, COUNT(*) AS bookmarks FROM post_bookmarks GROUP BY post_id), "
		query += "comments_agg AS (SELECT post_id, COUNT(*) AS comments FROM comments GROUP BY post_id) "

		query += "SELECT " // If duplicate records are returned, use SELECT DISTINCT on (posts.id) to remove duplicates instead of just SELECT
		query += "profiles.id AS profile_id, profiles.username AS profile_username, profiles.name AS profile_name, profiles.mini_avatar AS profile_mini_avatar, "
		query += "posts.id AS post_id, posts.title AS post_title, posts.caption AS post_caption, posts.created_at AS created_at, "
		query += "COALESCE(media_agg.media_data, '[]') AS media_data, "
		query += "COALESCE(likes_agg.likes, 0) AS num_likes, COALESCE(dislikes_agg.dislikes, 0) AS num_dislikes, COALESCE(bookmarks_agg.bookmarks, 0) AS num_bookmarks, COALESCE(comments_agg.comments, 0) AS num_comments, "
		query += "CASE WHEN pl.profile_id IS NOT NULL THEN true ELSE false END AS is_liked, "
		query += "CASE WHEN pd.profile_id IS NOT NULL THEN true ELSE false END AS is_disliked, "
		query += "true AS is_bookmarked "

		query += "FROM posts "
		query += "JOIN profiles ON posts.profile_id = profiles.id "
		query += "JOIN post_bookmarks ON post_bookmarks.post_id = posts.id "
		query += "LEFT JOIN media_agg ON posts.id = media_agg.post_id "
		query += "LEFT JOIN likes_agg ON posts.id = likes_agg.post_id "
		query += "LEFT JOIN dislikes_agg ON posts.id = dislikes_agg.post_id "
		query += "LEFT JOIN bookmarks_agg ON posts.id = bookmarks_agg.post_id "
		query += "LEFT JOIN comments_agg ON posts.id = comments_agg.post_id "
		query += "LEFT JOIN post_likes pl ON posts.id = pl.post_id AND pl.profile_id = ? "
		query += "LEFT JOIN post_dislikes pd ON posts.id = pd.post_id AND pd.profile_id = ? "

		query += "WHERE post_bookmarks.profile_id = ? AND posts.is_archived = false "
		query += "ORDER BY post_bookmarks.created_at DESC "
		query += "LIMIT ? OFFSET ?;"

		dbCtx, dbCancel := configs.NewQueryContext()
		defer dbCancel()
		errChan <- configs.Database.WithContext(dbCtx).Raw(query, reqProfile.Id, reqProfile.Id, reqProfile.Id, limit, offset).Scan(&bookmarkedPosts).Error
	}()

	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numBookmarks int64
	if err := configs.Database.WithContext(dbCtx2).Table("post_bookmarks").Where("profile_id = ?", reqProfile.Id).Count(&numBookmarks).Error; err != nil {
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
			"last_page":    int(math.Ceil(float64(numBookmarks) / float64(limit))),
			"data":         bookmarkedPosts,
		},
	}))
}
