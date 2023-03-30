package postcontrollers

import (
	"math"
	"strings"
	"sync"
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

	// Run both queries concurrently to reduce response time
	errChan := make(chan error, 1) // make this buffered so that the goroutine doesn't block
	wg := new(sync.WaitGroup)
	wg.Add(1)
	var bookmarkedPosts = []responses.Post{}
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
		query += "true AS is_bookmarked "
		query += "FROM posts "
		query += "JOIN profiles ON posts.profile_id = profiles.id "
		query += "JOIN post_bookmarks ON post_bookmarks.post_id = posts.id "
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
