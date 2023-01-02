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

type feedPosts struct {
	Id           string    `json:"id"`
	ProfileId    string    `json:"profile_id"`
	Username     string    `json:"profile_username"`
	Name         string    `json:"profile_name"`
	MiniAvatar   string    `json:"profile_avatar"`
	Title        string    `json:"title"`
	Caption      string    `json:"caption"`
	LikerId      string    `json:"-"`
	DislikerId   string    `json:"-"`
	BookmarkerId string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type feedPostsWithMedia struct {
	feedPosts
	DidLike     bool               `json:"did_like"`
	DidDislike  bool               `json:"did_dislike"`
	DidBookmark bool               `json:"did_bookmark"`
	Media       []models.PostMedia `json:"media"`
}

func GetFollowingsFeed(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	query := fmt.Sprintf(
		"SELECT post.id, post.created_at, post.title, post.caption, post.profile_id, profile.username, profile.name, profile.mini_avatar, post_likes.liker_id, post_dislikes.disliker_id, post_bookmarks.bookmarker_id"+
			" FROM profiles as profile"+
			" JOIN profile_followers"+
			" ON profile_followers.follower_id = \"%s\" AND profile_followers.profile_id = profile.id"+
			" JOIN posts as post"+
			" ON post.profile_id = profile.id AND post.is_archived = %d AND post.for_subscribers_only = %d"+
			" LEFT JOIN post_likes"+
			" ON post_likes.post_id = post.id AND post_likes.liker_id = \"%s\""+
			" LEFT JOIN post_dislikes"+
			" ON post_dislikes.post_id = post.id AND post_dislikes.disliker_id = \"%s\""+
			" LEFT JOIN post_bookmarks"+
			" ON post_bookmarks.post_id = post.id AND post_bookmarks.bookmarker_id = \"%s\""+
			" ORDER BY post.created_at DESC LIMIT %d OFFSET %d",
		reqProfile.Id, 0, 0, reqProfile.Id, reqProfile.Id, reqProfile.Id, limit, offset,
	)
	var followingFeedPosts = []feedPosts{}
	if err := configs.Database.Raw(query).Scan(&followingFeedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	var finalResult = []feedPostsWithMedia{}
	if len(followingFeedPosts) > 0 {
		// Create query to get the media objects of posts found in above query. Format is "SELECT * FROM post_media WHERE post_media.post_id IN (...postIds)"
		query = "SELECT * FROM post_media WHERE post_media.post_id IN ("
		for _, post := range followingFeedPosts {
			query += fmt.Sprintf("\"%s\",", post.Id)
			didLike, didDislike, didBookmark := false, false, false
			if post.LikerId != "" {
				didLike = true
			}
			if post.DislikerId != "" {
				didDislike = true
			}
			if post.BookmarkerId != "" {
				didBookmark = true
			}
			finalResult = append(finalResult, feedPostsWithMedia{feedPosts: post, DidLike: didLike, DidDislike: didDislike, DidBookmark: didBookmark})
		}
		query = strings.TrimSuffix(query, ",")
		query += ") ORDER BY created_at DESC"

		// Execute query
		var postMedia = []models.PostMedia{}
		if err := configs.Database.Raw(query).Scan(&postMedia).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
		}

		// Map media objects found to their proper post
		var i int = 0
		for _, media := range postMedia {
			for media.PostId != followingFeedPosts[i].Id {
				i++
			}
			finalResult[i].Media = append(finalResult[i].Media, media)
		}
	}

	// Get total number of feeds for post
	var numFollowingFeedPosts int
	query2 := fmt.Sprintf("SELECT count(*) FROM profiles as profile JOIN profile_followers ON profile_followers.follower_id = \"%s\" AND profile_followers.profile_id = profile.id JOIN posts as post ON post.profile_id = profile.id WHERE post.is_archived = %d AND post.for_subscribers_only = %d", reqProfile.Id, 0, 0)
	if err := configs.Database.Raw(query2).Scan(&numFollowingFeedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numFollowingFeedPosts) / float64(limit))),
			"data":         finalResult,
		},
	}))
}

func GetSubscriptionsFeed(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	query := fmt.Sprintf(
		"SELECT post.id, post.created_at, post.title, post.caption, post.profile_id, profile.username, profile.name, profile.mini_avatar, post_likes.liker_id, post_dislikes.disliker_id, post_bookmarks.bookmarker_id"+
			" FROM profiles as profile"+
			" JOIN profile_subscribers"+
			" ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profile.id"+
			" JOIN posts as post"+
			" ON post.profile_id = profile.id AND post.is_archived = %d AND post.for_subscribers_only = %d"+
			" LEFT JOIN post_likes"+
			" ON post_likes.post_id = post.id AND post_likes.liker_id = \"%s\""+
			" LEFT JOIN post_dislikes"+
			" ON post_dislikes.post_id = post.id AND post_dislikes.disliker_id = \"%s\""+
			" LEFT JOIN post_bookmarks"+
			" ON post_bookmarks.post_id = post.id AND post_bookmarks.bookmarker_id = \"%s\""+
			" ORDER BY post.created_at DESC LIMIT %d OFFSET %d",
		reqProfile.Id, 0, 1, reqProfile.Id, reqProfile.Id, reqProfile.Id, limit, offset,
	)
	var subscriptionsFeedPosts = []feedPosts{}
	if err := configs.Database.Raw(query).Scan(&subscriptionsFeedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	var finalResult = []feedPostsWithMedia{}
	if len(subscriptionsFeedPosts) > 0 {
		// Create query to get the media objects of posts found in above query. Format is "SELECT * FROM post_media WHERE post_media.post_id IN (...postIds)"
		query = "SELECT * FROM post_media WHERE post_media.post_id IN ("
		for _, post := range subscriptionsFeedPosts {
			query += fmt.Sprintf("\"%s\",", post.Id)
			didLike, didDislike, didBookmark := false, false, false
			if post.LikerId != "" {
				didLike = true
			}
			if post.DislikerId != "" {
				didDislike = true
			}
			if post.BookmarkerId != "" {
				didBookmark = true
			}
			finalResult = append(finalResult, feedPostsWithMedia{feedPosts: post, DidLike: didLike, DidDislike: didDislike, DidBookmark: didBookmark})
		}
		query = strings.TrimSuffix(query, ",")
		query += ") ORDER BY created_at DESC"

		// Execute query
		var postMedia = []models.PostMedia{}
		if err := configs.Database.Raw(query).Scan(&postMedia).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
		}

		// Map media objects found to their proper post
		var i int = 0
		for _, media := range postMedia {
			for media.PostId != subscriptionsFeedPosts[i].Id {
				i++
			}
			finalResult[i].Media = append(finalResult[i].Media, media)
		}
	}

	// Get total number of feeds for post
	var numSubscriptionsFeedPosts int
	query2 := fmt.Sprintf("SELECT count(*) FROM profiles as profile JOIN profile_subscribers ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profile.id JOIN posts as post ON post.profile_id = profile.id WHERE post.is_archived = %d AND post.for_subscribers_only = %d", reqProfile.Id, 0, 1)
	if err := configs.Database.Raw(query2).Scan(&numSubscriptionsFeedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numSubscriptionsFeedPosts) / float64(limit))),
			"data":         finalResult,
		},
	}))
}

func GetArchivedPosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Get archived posts(paginated)
	var archivedPosts = []models.Post{}
	if err := configs.Database.Model(&reqProfile).Offset(offset).Limit(limit).Order("posts.created_at DESC").Where("is_archived = ?", true).Preload("Media").Association("Posts").Find(&archivedPosts); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of archived posts
	numArchivedPosts := configs.Database.Model(&reqProfile).Where("is_archived = ?", true).Association("Posts").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numArchivedPosts) / float64(limit))),
			"data":         archivedPosts,
		},
	}))
}

func GetPublicPosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	// Get public posts(paginated)
	var publicPosts = []models.Post{}
	if err := configs.Database.Model(&models.Post{}).Preload("Media").Offset(offset).Limit(limit).Order("created_at DESC").Find(&publicPosts, "profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of public posts
	var numPublicPosts int64
	if err := configs.Database.Model(&models.Post{}).Where("profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, false).Count(&numPublicPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numPublicPosts) / float64(limit))),
			"data":         publicPosts,
		},
	}))
}

func GetExclusivePosts(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)

	// Get exclusive posts(paginated)
	var exclusivePosts = []models.Post{}
	if err := configs.Database.Model(&models.Post{}).Preload("Media").Offset(offset).Limit(limit).Order("created_at DESC").Find(&exclusivePosts, "profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, true).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of exclusive posts
	var numExclusivePosts int64
	if err := configs.Database.Model(&models.Post{}).Where("profile_id = ? AND is_archived = ? AND for_subscribers_only = ?", c.Params("profileId"), false, true).Count(&numExclusivePosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numExclusivePosts) / float64(limit))),
			"data":         exclusivePosts,
		},
	}))
}
