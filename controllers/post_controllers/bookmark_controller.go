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

type postsWithoutMedia struct {
	PostId    string    `json:"post_id"`
	Title     string    `json:"title" gorm:"column:post_title"`
	Caption   string    `json:"caption" gorm:"column:post_caption"`
	CreatedAt time.Time `json:"created_at"`

	ProfileId  string `json:"profile_id"`
	Username   string `json:"username" gorm:"column:profile_username"`
	Name       string `json:"name" gorm:"column:profile_name"`
	MiniAvatar string `json:"mini_avatar" gorm:"column:profile_mini_avatar"`

	MediaUrls string `json:"-" gorm:"column:media_urls"`
	IsImages  string `json:"-" gorm:"column:is_images"`
	IsVideos  string `json:"-" gorm:"column:is_videos"`
	IsAudios  string `json:"-" gorm:"column:is_audios"`

	NumLikes     int `json:"num_likes"`
	NumDislikes  int `json:"num_dislikes"`
	NumBookmarks int `json:"num_bookmarks"`

	IsLiked      bool `json:"is_liked"`
	IsDisliked   bool `json:"is_disliked"`
	IsBookmarked bool `json:"is_bookmarked"`
}

type miniPostMedia struct {
	MediaUrl string `json:"media_url"`
	IsImage  bool   `json:"is_image"`
	IsVideo  bool   `json:"is_video"`
	IsAudio  bool   `json:"is_audio"`
}

type postsWithMedia struct {
	postsWithoutMedia
	Media []miniPostMedia `json:"media"`
}

func BookmarkPost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	newBookmarkObj := models.PostBookmark{
		PostId:    c.Params("postId"),
		ProfileId: reqProfile.Id,
		CreatedAt: time.Now(),
	}
	if err := configs.Database.Table("post_bookmarks").Where("post_id = ? AND profile_id = ?", newBookmarkObj.PostId, newBookmarkObj.ProfileId).FirstOrCreate(&newBookmarkObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been bookmarked."}))
}

func RemoveBookmark(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	var bookmarkObj models.PostBookmark
	if err := configs.Database.Table("post_bookmarks").Delete(&bookmarkObj, "post_id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Bookmark has been removed."}))
}

/*
   b => post_bookmarks
   p => posts
   u => profiles
   m => post_media
   l => post_likes
   d => post_dislikes
*/

func GetBookmarkedPosts(c *fiber.Ctx) error {
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
			"l.num_likes, d.num_dislikes, b2.num_bookmarks, "+
			"(CASE WHEN (SELECT profile_id FROM post_likes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_liked, "+
			"(CASE WHEN (SELECT profile_id FROM post_dislikes WHERE post_id = p.id AND profile_id = \"%s\") IS NOT NULL THEN 1 ELSE 0 END) AS is_disliked "+
			"FROM post_bookmarks b "+
			"JOIN posts p ON b.post_id = p.id AND b.profile_id = \"%s\" AND p.is_archived = %d "+
			"JOIN profiles u ON p.profile_id = u.id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_likes FROM post_likes GROUP by post_id) l ON p.id = l.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_dislikes FROM post_dislikes GROUP by post_id) d ON p.id = d.post_id "+
			"LEFT JOIN (SELECT post_id, COUNT(*) AS num_bookmarks FROM post_bookmarks GROUP by post_id) b2 ON p.id = b2.post_id "+
			"GROUP BY p.id, u.id "+
			"ORDER BY b.created_at DESC "+
			"LIMIT %d OFFSET %d",
		reqProfile.Id, reqProfile.Id, reqProfile.Id, 0, limit, offset,
	)
	var unpreparedBookmarkedPosts = []postsWithoutMedia{}
	if err := configs.Database.Raw(query).Scan(&unpreparedBookmarkedPosts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	var bookmarkedPosts = preparePosts(&unpreparedBookmarkedPosts, true, false, false)

	var numBookmarks int64
	if err := configs.Database.Table("post_bookmarks").Where("profile_id = ?", reqProfile.Id).Count(&numBookmarks).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numBookmarks) / float64(limit))),
			"data":         bookmarkedPosts,
		},
	}))
}

/*
Takes in an array of posts without their media objects properly structured and returns an array with properly structured media objects.
Also takes in three booleans which mark each post as the variable name suggests. This is for when we know for a fact that a certain batch of posts is bookmarked/liked/disliked
*/
func preparePosts(unpreparedPosts *[]postsWithoutMedia, markPostsAsBookmarked bool, markPostsAsLiked bool, markPostsAsDisliked bool) []postsWithMedia {
	var preppedPosts = []postsWithMedia{}

	for _, post := range *unpreparedPosts { // The limit of results of paginated data is capped at 25 and the post_media for a post is capped at 5. Therefore this loop has at most 125 iterations
		if markPostsAsBookmarked {
			post.IsBookmarked = true
		}
		if markPostsAsLiked {
			post.IsLiked = true
		}
		if markPostsAsDisliked {
			post.IsDisliked = true
		}
		var mediaObjs = []miniPostMedia{}
		mediaUrls := strings.Split(post.MediaUrls, ",")
		isImages := strings.Split(post.IsImages, ",")
		isVideos := strings.Split(post.IsVideos, ",")
		isAudios := strings.Split(post.IsAudios, ",")
		for i, url := range mediaUrls {
			if len(post.MediaUrls) == 0 { // len(strings.Split("", ",")) == 1 because the Split returns [] with empty string inside it meaning this loop will run. this statement breaks the loop
				break
			}
			mediaObjs = append(mediaObjs, miniPostMedia{MediaUrl: url, IsImage: stringToBool(isImages[i]), IsVideo: stringToBool(isVideos[i]), IsAudio: stringToBool(isAudios[i])})
		}
		preppedPosts = append(preppedPosts, postsWithMedia{postsWithoutMedia: post, Media: mediaObjs})
	}

	return preppedPosts
}

func stringToBool(s string) bool {
	if s == "1" {
		return true
	} else {
		return false
	}
}
