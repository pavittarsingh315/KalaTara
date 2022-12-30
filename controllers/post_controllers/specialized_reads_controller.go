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
	Id         string    `json:"id"`
	ProfileId  string    `json:"profile_id"`
	Username   string    `json:"profile_username"`
	Name       string    `json:"profile_name"`
	MiniAvatar string    `json:"profile_avatar"`
	Title      string    `json:"title"`
	Caption    string    `json:"caption"`
	CreatedAt  time.Time `json:"created_at"`
}

type feedPostsWithMedia struct {
	feedPosts
	Media []models.PostMedia `json:"media"`
}

func GetFollowingsFeed(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	query := fmt.Sprintf(
		"SELECT post.id, post.created_at, post.title, post.caption, post.profile_id, profile.username, profile.name, profile.mini_avatar"+
			" FROM profiles as profile"+
			" JOIN profile_followers"+
			" ON profile_followers.follower_id = \"%s\" AND profile_followers.profile_id = profile.id"+
			" JOIN posts as post"+
			" ON post.profile_id = profile.id"+
			" WHERE post.is_archived = %d AND post.for_subscribers_only = %d"+
			" ORDER BY post.created_at DESC LIMIT %d OFFSET %d",
		reqProfile.Id, 0, 0, limit, offset,
	)
	var followingFeedPosts = []feedPosts{}
	if err := configs.Database.Raw(query).Scan(&followingFeedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	var finalResult = []feedPostsWithMedia{}
	// SELECT * FROM `post_media` WHERE `post_media`.`post_id` IN (...postIds)
	query = "SELECT * FROM post_media WHERE post_media.post_id IN ("
	for _, post := range followingFeedPosts {
		query += fmt.Sprintf("\"%s\",", post.Id)
		finalResult = append(finalResult, feedPostsWithMedia{feedPosts: post})
	}
	query = strings.TrimSuffix(query, ",")
	query += ")"

	var postMedia = []models.PostMedia{}
	if err := configs.Database.Raw(query).Scan(&postMedia).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// TODO: if any time in the future posts are allowed to be made without media being required, you'll need to update this logic which maps the post media to its proper post
	var i int = 0
	var lastPostId string = followingFeedPosts[0].Id
	for _, media := range postMedia {
		if media.PostId != lastPostId {
			i++
		}
		finalResult[i].Media = append(finalResult[i].Media, media)
		lastPostId = media.PostId
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
		"SELECT post.id, post.created_at, post.title, post.caption, post.profile_id, profile.username, profile.name, profile.mini_avatar"+
			" FROM profiles as profile"+
			" JOIN profile_subscribers"+
			" ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profile.id"+
			" JOIN posts as post"+
			" ON post.profile_id = profile.id"+
			" WHERE post.is_archived = %d AND post.for_subscribers_only = %d"+
			" ORDER BY post.created_at DESC LIMIT %d OFFSET %d",
		reqProfile.Id, 0, 1, limit, offset,
	)
	var subscriptionsFeedPosts = []feedPosts{}
	if err := configs.Database.Raw(query).Scan(&subscriptionsFeedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	var finalResult = []feedPostsWithMedia{}
	// SELECT * FROM `post_media` WHERE `post_media`.`post_id` IN (...postIds)
	query = "SELECT * FROM post_media WHERE post_media.post_id IN ("
	for _, post := range subscriptionsFeedPosts {
		query += fmt.Sprintf("\"%s\",", post.Id)
		finalResult = append(finalResult, feedPostsWithMedia{feedPosts: post})
	}
	query = strings.TrimSuffix(query, ",")
	query += ")"

	var postMedia = []models.PostMedia{}
	if err := configs.Database.Raw(query).Scan(&postMedia).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// TODO: if any time in the future posts are allowed to be made without media being required, you'll need to update this logic which maps the post media to its proper post
	var i int = 0
	var lastPostId string = subscriptionsFeedPosts[0].Id
	for _, media := range postMedia {
		if media.PostId != lastPostId {
			i++
		}
		finalResult[i].Media = append(finalResult[i].Media, media)
		lastPostId = media.PostId
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
