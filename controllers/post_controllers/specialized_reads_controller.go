package postcontrollers

import (
	"fmt"
	"math"
	"sync"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func GetFollowingsFeed(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Run both queries concurrently to reduce response time
	errChan := make(chan error, 1) // make this buffered so that the goroutine doesn't block
	wg := new(sync.WaitGroup)
	wg.Add(1)
	var feedPosts = []responses.Post{}
	go func() {
		defer wg.Done()

		query := "SELECT profiles.id AS profile_id, profiles.username AS profile_username, profiles.name AS profile_name, profiles.mini_avatar AS profile_mini_avatar, "
		query += "posts.id AS post_id, posts.title AS post_title, posts.caption AS post_caption, posts.created_at AS created_at, "
		query += "(SELECT json_agg(json_build_object('media_url', m.media_url, 'is_image', m.is_image, 'is_video', m.is_video, 'is_audio', m.is_audio) ORDER BY m.position) FROM post_media m WHERE m.post_id = posts.id) AS media_data, "
		query += "(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) AS num_likes, "
		query += "(SELECT COUNT(*) FROM post_dislikes WHERE post_dislikes.post_id = posts.id) AS num_dislikes, "
		query += "(SELECT COUNT(*) FROM post_bookmarks WHERE post_bookmarks.post_id = posts.id) AS num_bookmarks, "
		query += "EXISTS(SELECT 1 FROM post_likes WHERE post_likes.post_id = posts.id AND post_likes.profile_id = ?) AS is_liked, "
		query += "EXISTS(SELECT 1 FROM post_dislikes WHERE post_dislikes.post_id = posts.id AND post_dislikes.profile_id = ?) AS is_disliked, "
		query += "EXISTS(SELECT 1 FROM post_bookmarks WHERE post_bookmarks.post_id = posts.id AND post_bookmarks.profile_id = ?) AS is_bookmarked "
		query += "FROM profile_followers "
		query += "JOIN posts ON posts.profile_id = profile_followers.profile_id "
		query += "JOIN profiles ON profiles.id = profile_followers.profile_id "
		query += "WHERE profile_followers.follower_id = ? AND posts.is_archived = false AND posts.for_subscribers_only = false "
		query += "ORDER BY posts.created_at DESC "
		query += "LIMIT ? OFFSET ?;"

		dbCtx, dbCancel := configs.NewQueryContext()
		defer dbCancel()
		errChan <- configs.Database.WithContext(dbCtx).Raw(query, reqProfile.Id, reqProfile.Id, reqProfile.Id, reqProfile.Id, limit, offset).Scan(&feedPosts).Error
	}()

	// Get total number of posts in following feed
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numFeedPosts int64
	if err := configs.Database.WithContext(dbCtx2).Table("posts").Where("profile_id IN (SELECT profile_id FROM profile_followers WHERE follower_id = ?) AND is_archived = ? AND for_subscribers_only = ?", reqProfile.Id, false, false).Count(&numFeedPosts).Error; err != nil {
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
			"last_page":    int(math.Ceil(float64(numFeedPosts) / float64(limit))),
			"data":         feedPosts,
		},
	}))
}

func GetSubscriptionsFeed(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Run both queries concurrently to reduce response time
	errChan := make(chan error, 1) // make this buffered so that the goroutine doesn't block
	wg := new(sync.WaitGroup)
	wg.Add(1)
	var feedPosts = []responses.Post{}
	go func() {
		defer wg.Done()

		query := "SELECT profiles.id AS profile_id, profiles.username AS profile_username, profiles.name AS profile_name, profiles.mini_avatar AS profile_mini_avatar, "
		query += "posts.id AS post_id, posts.title AS post_title, posts.caption AS post_caption, posts.created_at AS created_at, "
		query += "(SELECT json_agg(json_build_object('media_url', m.media_url, 'is_image', m.is_image, 'is_video', m.is_video, 'is_audio', m.is_audio) ORDER BY m.position) FROM post_media m WHERE m.post_id = posts.id) AS media_data, "
		query += "(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) AS num_likes, "
		query += "(SELECT COUNT(*) FROM post_dislikes WHERE post_dislikes.post_id = posts.id) AS num_dislikes, "
		query += "(SELECT COUNT(*) FROM post_bookmarks WHERE post_bookmarks.post_id = posts.id) AS num_bookmarks, "
		query += "EXISTS(SELECT 1 FROM post_likes WHERE post_likes.post_id = posts.id AND post_likes.profile_id = ?) AS is_liked, "
		query += "EXISTS(SELECT 1 FROM post_dislikes WHERE post_dislikes.post_id = posts.id AND post_dislikes.profile_id = ?) AS is_disliked, "
		query += "EXISTS(SELECT 1 FROM post_bookmarks WHERE post_bookmarks.post_id = posts.id AND post_bookmarks.profile_id = ?) AS is_bookmarked "
		query += "FROM profile_subscribers "
		query += "JOIN posts ON posts.profile_id = profile_subscribers.profile_id "
		query += "JOIN profiles ON profiles.id = profile_subscribers.profile_id "
		query += "WHERE profile_subscribers.subscriber_id = ? AND profile_subscribers.is_accepted = true AND posts.is_archived = false AND posts.for_subscribers_only = true "
		query += "ORDER BY posts.created_at DESC "
		query += "LIMIT ? OFFSET ?;"

		dbCtx, dbCancel := configs.NewQueryContext()
		defer dbCancel()
		errChan <- configs.Database.WithContext(dbCtx).Raw(query, reqProfile.Id, reqProfile.Id, reqProfile.Id, reqProfile.Id, limit, offset).Scan(&feedPosts).Error
	}()

	// Get total number of following feed posts
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numFeedPosts int64
	if err := configs.Database.WithContext(dbCtx2).Table("posts").Where("profile_id IN (SELECT profile_id FROM profile_subscribers WHERE subscriber_id = ? AND is_accepted = true) AND is_archived = ? AND for_subscribers_only = ?", reqProfile.Id, false, true).Count(&numFeedPosts).Error; err != nil {
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
			"last_page":    int(math.Ceil(float64(numFeedPosts) / float64(limit))),
			"data":         feedPosts,
		},
	}))
}

func GetArchivedPosts(c *fiber.Ctx) error {
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
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Raw(query).Scan(&unpreparedArchivedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	var archivedPosts = preparePosts(&unpreparedArchivedPosts, false, false, false)

	// Get total number of archived posts
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	numArchivedPosts := configs.Database.WithContext(dbCtx2).Model(&reqProfile).Where("is_archived = ?", true).Association("Posts").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numArchivedPosts) / float64(limit))),
			"data":         archivedPosts,
		},
	}))
}

func GetPublicPosts(c *fiber.Ctx) error {
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
			"l.num_likes, d.num_dislikes, b.num_bookmarks, "+
			"(CASE WHEN (SELECT profile_id FROM post_likes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_liked, "+
			"(CASE WHEN (SELECT profile_id FROM post_dislikes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_disliked, "+
			"(CASE WHEN (SELECT profile_id FROM post_bookmarks WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_bookmarked "+
			"FROM posts p "+
			"JOIN profiles u ON p.profile_id = u.id AND u.id = \"%s\" AND p.is_archived = %d AND p.for_subscribers_only = %d "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_likes FROM post_likes GROUP by post_id) l ON p.id = l.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_dislikes FROM post_dislikes GROUP by post_id) d ON p.id = d.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_bookmarks FROM post_bookmarks GROUP by post_id) b ON p.id = b.post_id "+
			"GROUP BY p.id, u.id "+
			"ORDER BY p.created_at DESC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, reqProfile.Id, c.Params("profileId"), 0, 0, limit, offset,
	)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var unpreparedPublicPosts = []postsWithoutMedia{}
	if err := configs.Database.WithContext(dbCtx).Raw(query).Scan(&unpreparedPublicPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	var publicPosts = preparePosts(&unpreparedPublicPosts, false, false, false)

	// Get total number of public posts
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numPublicPosts int64
	if err := configs.Database.WithContext(dbCtx2).Model(&models.Post{}).Where("profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, false).Count(&numPublicPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numPublicPosts) / float64(limit))),
			"data":         publicPosts,
		},
	}))
}

func GetExclusivePosts(c *fiber.Ctx) error {
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
			"l.num_likes, d.num_dislikes, b.num_bookmarks, "+
			"(CASE WHEN (SELECT profile_id FROM post_likes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_liked, "+
			"(CASE WHEN (SELECT profile_id FROM post_dislikes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_disliked, "+
			"(CASE WHEN (SELECT profile_id FROM post_bookmarks WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_bookmarked "+
			"FROM posts p "+
			"JOIN profiles u ON p.profile_id = u.id AND u.id = \"%s\" AND p.is_archived = %d AND p.for_subscribers_only = %d "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_likes FROM post_likes GROUP by post_id) l ON p.id = l.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_dislikes FROM post_dislikes GROUP by post_id) d ON p.id = d.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_bookmarks FROM post_bookmarks GROUP by post_id) b ON p.id = b.post_id "+
			"GROUP BY p.id, u.id "+
			"ORDER BY p.created_at DESC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, reqProfile.Id, c.Params("profileId"), 0, 1, limit, offset,
	)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var unpreparedExclusivePosts = []postsWithoutMedia{}
	if err := configs.Database.WithContext(dbCtx).Raw(query).Scan(&unpreparedExclusivePosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	var exclusivePosts = preparePosts(&unpreparedExclusivePosts, false, false, false)

	// Get total number of public posts
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numExclusivePosts int64
	if err := configs.Database.WithContext(dbCtx2).Model(&models.Post{}).Where("profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, true).Count(&numExclusivePosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numExclusivePosts) / float64(limit))),
			"data":         exclusivePosts,
		},
	}))
}
