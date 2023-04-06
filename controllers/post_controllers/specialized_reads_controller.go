package postcontrollers

import (
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
		query += "CASE WHEN pb.profile_id IS NOT NULL THEN true ELSE false END AS is_bookmarked "

		query += "FROM profile_followers "
		query += "JOIN posts ON posts.profile_id = profile_followers.profile_id "
		query += "JOIN profiles ON profiles.id = profile_followers.profile_id "
		query += "LEFT JOIN media_agg ON posts.id = media_agg.post_id "
		query += "LEFT JOIN likes_agg ON posts.id = likes_agg.post_id "
		query += "LEFT JOIN dislikes_agg ON posts.id = dislikes_agg.post_id "
		query += "LEFT JOIN bookmarks_agg ON posts.id = bookmarks_agg.post_id "
		query += "LEFT JOIN comments_agg ON posts.id = comments_agg.post_id "
		query += "LEFT JOIN post_likes pl ON posts.id = pl.post_id AND pl.profile_id = ? "
		query += "LEFT JOIN post_dislikes pd ON posts.id = pd.post_id AND pd.profile_id = ? "
		query += "LEFT JOIN post_bookmarks pb ON posts.id = pb.post_id AND pb.profile_id = ? "

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
		query += "CASE WHEN pb.profile_id IS NOT NULL THEN true ELSE false END AS is_bookmarked "

		query += "FROM profile_subscribers "
		query += "JOIN posts ON posts.profile_id = profile_subscribers.profile_id "
		query += "JOIN profiles ON profiles.id = profile_subscribers.profile_id "
		query += "LEFT JOIN media_agg ON posts.id = media_agg.post_id "
		query += "LEFT JOIN likes_agg ON posts.id = likes_agg.post_id "
		query += "LEFT JOIN dislikes_agg ON posts.id = dislikes_agg.post_id "
		query += "LEFT JOIN bookmarks_agg ON posts.id = bookmarks_agg.post_id "
		query += "LEFT JOIN comments_agg ON posts.id = comments_agg.post_id "
		query += "LEFT JOIN post_likes pl ON posts.id = pl.post_id AND pl.profile_id = ? "
		query += "LEFT JOIN post_dislikes pd ON posts.id = pd.post_id AND pd.profile_id = ? "
		query += "LEFT JOIN post_bookmarks pb ON posts.id = pb.post_id AND pb.profile_id = ? "

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

	// Run both queries concurrently to reduce response time
	errChan := make(chan error, 1) // make this buffered so that the goroutine doesn't block
	wg := new(sync.WaitGroup)
	wg.Add(1)
	var archivedPosts = []responses.Post{}
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
		query += "CASE WHEN pb.profile_id IS NOT NULL THEN true ELSE false END AS is_bookmarked "

		query += "FROM profiles "
		query += "JOIN posts ON posts.profile_id = profiles.id "
		query += "LEFT JOIN media_agg ON posts.id = media_agg.post_id "
		query += "LEFT JOIN likes_agg ON posts.id = likes_agg.post_id "
		query += "LEFT JOIN dislikes_agg ON posts.id = dislikes_agg.post_id "
		query += "LEFT JOIN bookmarks_agg ON posts.id = bookmarks_agg.post_id "
		query += "LEFT JOIN comments_agg ON posts.id = comments_agg.post_id "
		query += "LEFT JOIN post_likes pl ON posts.id = pl.post_id AND pl.profile_id = ? "
		query += "LEFT JOIN post_dislikes pd ON posts.id = pd.post_id AND pd.profile_id = ? "
		query += "LEFT JOIN post_bookmarks pb ON posts.id = pb.post_id AND pb.profile_id = ? "

		query += "WHERE profiles.id = ? AND posts.is_archived = true "
		query += "ORDER BY posts.created_at DESC "
		query += "LIMIT ? OFFSET ?;"

		dbCtx, dbCancel := configs.NewQueryContext()
		defer dbCancel()
		errChan <- configs.Database.WithContext(dbCtx).Raw(query, reqProfile.Id, reqProfile.Id, reqProfile.Id, reqProfile.Id, limit, offset).Scan(&archivedPosts).Error
	}()

	// Get total number of archived posts
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	numArchivedPosts := configs.Database.WithContext(dbCtx2).Model(&reqProfile).Where("is_archived = ?", true).Association("Posts").Count()

	wg.Wait()

	if err := <-errChan; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	close(errChan)

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

	// Run both queries concurrently to reduce response time
	errChan := make(chan error, 1) // make this buffered so that the goroutine doesn't block
	wg := new(sync.WaitGroup)
	wg.Add(1)
	var publicPosts = []responses.Post{}
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
		query += "CASE WHEN pb.profile_id IS NOT NULL THEN true ELSE false END AS is_bookmarked "

		query += "FROM profiles "
		query += "JOIN posts ON posts.profile_id = profiles.id "
		query += "LEFT JOIN media_agg ON posts.id = media_agg.post_id "
		query += "LEFT JOIN likes_agg ON posts.id = likes_agg.post_id "
		query += "LEFT JOIN dislikes_agg ON posts.id = dislikes_agg.post_id "
		query += "LEFT JOIN bookmarks_agg ON posts.id = bookmarks_agg.post_id "
		query += "LEFT JOIN comments_agg ON posts.id = comments_agg.post_id "
		query += "LEFT JOIN post_likes pl ON posts.id = pl.post_id AND pl.profile_id = ? "
		query += "LEFT JOIN post_dislikes pd ON posts.id = pd.post_id AND pd.profile_id = ? "
		query += "LEFT JOIN post_bookmarks pb ON posts.id = pb.post_id AND pb.profile_id = ? "

		query += "WHERE profiles.id = ? AND posts.is_archived = false AND posts.for_subscribers_only = false "
		query += "ORDER BY posts.created_at DESC "
		query += "LIMIT ? OFFSET ?;"

		dbCtx, dbCancel := configs.NewQueryContext()
		defer dbCancel()
		errChan <- configs.Database.WithContext(dbCtx).Raw(query, reqProfile.Id, reqProfile.Id, reqProfile.Id, c.Params("profileId"), limit, offset).Scan(&publicPosts).Error
	}()

	// Get total number of public posts
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numPublicPosts int64
	if err := configs.Database.WithContext(dbCtx2).Table("posts").Where("profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, false).Count(&numPublicPosts).Error; err != nil {
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

	// Run both queries concurrently to reduce response time
	errChan := make(chan error, 1) // make this buffered so that the goroutine doesn't block
	wg := new(sync.WaitGroup)
	wg.Add(1)
	var exclusivePosts = []responses.Post{}
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
		query += "CASE WHEN pb.profile_id IS NOT NULL THEN true ELSE false END AS is_bookmarked "

		query += "FROM profiles "
		query += "JOIN posts ON posts.profile_id = profiles.id "
		query += "LEFT JOIN media_agg ON posts.id = media_agg.post_id "
		query += "LEFT JOIN likes_agg ON posts.id = likes_agg.post_id "
		query += "LEFT JOIN dislikes_agg ON posts.id = dislikes_agg.post_id "
		query += "LEFT JOIN bookmarks_agg ON posts.id = bookmarks_agg.post_id "
		query += "LEFT JOIN comments_agg ON posts.id = comments_agg.post_id "
		query += "LEFT JOIN post_likes pl ON posts.id = pl.post_id AND pl.profile_id = ? "
		query += "LEFT JOIN post_dislikes pd ON posts.id = pd.post_id AND pd.profile_id = ? "
		query += "LEFT JOIN post_bookmarks pb ON posts.id = pb.post_id AND pb.profile_id = ? "

		query += "WHERE profiles.id = ? AND posts.is_archived = false AND posts.for_subscribers_only = true "
		query += "ORDER BY posts.created_at DESC "
		query += "LIMIT ? OFFSET ?;"

		dbCtx, dbCancel := configs.NewQueryContext()
		defer dbCancel()
		errChan <- configs.Database.WithContext(dbCtx).Raw(query, reqProfile.Id, reqProfile.Id, reqProfile.Id, c.Params("profileId"), limit, offset).Scan(&exclusivePosts).Error
	}()

	// Get total number of exclusive posts
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numExclusivePosts int64
	if err := configs.Database.WithContext(dbCtx2).Table("posts").Where("profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, true).Count(&numExclusivePosts).Error; err != nil {
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
			"last_page":    int(math.Ceil(float64(numExclusivePosts) / float64(limit))),
			"data":         exclusivePosts,
		},
	}))
}
