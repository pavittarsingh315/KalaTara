package postcontrollers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rivo/uniseg"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

type mediaBody struct {
	MediaUrl string `json:"media_url"`
	IsImage  *bool  `json:"is_image"`
	IsVideo  *bool  `json:"is_video"`
	IsAudio  *bool  `json:"is_audio"`
}

func CreatePost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)
	reqBody := struct {
		Title              string      `json:"title"`
		Caption            string      `json:"caption"`
		ForSubscribersOnly *bool       `json:"for_subscribers_only"`
		IsArchived         *bool       `json:"is_archived"`
		Media              []mediaBody `json:"media"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	if reqBody.Title == "" || reqBody.Caption == "" || len(reqBody.Media) == 0 || reqBody.IsArchived == nil || reqBody.ForSubscribersOnly == nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}
	if len(reqBody.Media) > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Max number of attachments is 5."}))
	}
	for index, mediaObj := range reqBody.Media {
		if mediaObj.MediaUrl == "" || mediaObj.IsImage == nil || mediaObj.IsAudio == nil || mediaObj.IsVideo == nil {
			errMessage := fmt.Sprintf("Media entry #%d does not include all fields.", index+1)
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": errMessage}))
		}
		// this if block reads "if mediaObj is not just an image AND its not just a video AND its not just an audio, then either all fields are false or more than one field is true"
		if !(*mediaObj.IsImage && !*mediaObj.IsVideo && !*mediaObj.IsAudio) && !(!*mediaObj.IsImage && *mediaObj.IsVideo && !*mediaObj.IsAudio) && !(!*mediaObj.IsImage && !*mediaObj.IsVideo && *mediaObj.IsAudio) {
			errMessage := fmt.Sprintf("One and only one of the following media object fields must be true: is_image, is_audio, is_video. Media entry #%d violates this.", index+1)
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": errMessage}))
		}
	}

	reqBody.Title = strings.TrimSpace(reqBody.Title)     // remove leading and trailing whitespace
	reqBody.Caption = strings.TrimSpace(reqBody.Caption) // remove leading and trailing whitespace

	titleLength := uniseg.GraphemeClusterCount(reqBody.Title)
	captionLength := uniseg.GraphemeClusterCount(reqBody.Caption)
	if titleLength > 150 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Title is too long."}))
	}
	if captionLength > 200 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Caption is too long."}))
	}

	newPost := models.Post{
		ProfileId:          reqProfile.Id,
		Title:              reqBody.Title,
		Caption:            reqBody.Caption,
		ForSubscribersOnly: *reqBody.ForSubscribersOnly,
		IsArchived:         *reqBody.IsArchived,
	}
	if err := configs.Database.Model(&models.Post{}).Create(&newPost).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	var postMedia = []models.PostMedia{}
	for index, mediaObj := range reqBody.Media {
		postMedia = append(postMedia, models.PostMedia{
			PostId:   newPost.Id,
			Position: index,
			MediaUrl: mediaObj.MediaUrl,
			IsImage:  *mediaObj.IsImage,
			IsVideo:  *mediaObj.IsVideo,
			IsAudio:  *mediaObj.IsAudio,
		})
	}
	if err := configs.Database.Model(&models.PostMedia{}).Create(&postMedia).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	newPost.Media = postMedia

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": newPost}))
}

func GetPost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var post models.Post
	if err := configs.Database.Model(&models.Post{}).Preload("Media").Find(&post, "id = ?", c.Params("postId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if post.Id == "" { // Id field is empty => post does not exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This post does not exist."}))
	}

	// if request user is owner, return the post because its the owner
	if post.ProfileId == reqProfile.Id {
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": post}))
	} else if !post.IsArchived && !post.ForSubscribersOnly { // if post is not archived and is not hidden, return it
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": post}))
	}

	/*
	   At this point, the possible values for post.IsArchived and post.ForSubscribersOnly is
	      1. True, True
	      2. True, False
	      3. False, True
	*/

	if post.IsArchived {
		return c.Status(fiber.StatusLocked).JSON(responses.NewErrorResponse(fiber.StatusLocked, &fiber.Map{"data": "You do not have access to this post."}))
	}

	// At this point, we know post.IsArchived is false so post.ForSubscribersOnly must be true

	// Check if request user is subscribed to post owner
	var subscriberObj models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriberObj, "profile_id = ? AND subscriber_id = ?", post.ProfileId, reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if subscriberObj.ProfileId != "" && subscriberObj.SubscriberId != "" { // if both fields are populated, request user is subscribed to post owner
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": post}))
	} else {
		return c.Status(fiber.StatusLocked).JSON(responses.NewErrorResponse(fiber.StatusLocked, &fiber.Map{"data": "This post is limited to subscribers only."}))
	}
}

func EditPost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)
	reqBody := struct {
		Title      string `json:"title"`
		Caption    string `json:"caption"`
		IsArchived *bool  `json:"is_archived"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	if reqBody.Title == "" || reqBody.Caption == "" || reqBody.IsArchived == nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}

	reqBody.Title = strings.TrimSpace(reqBody.Title)     // remove leading and trailing whitespace
	reqBody.Caption = strings.TrimSpace(reqBody.Caption) // remove leading and trailing whitespace

	titleLength := uniseg.GraphemeClusterCount(reqBody.Title)
	captionLength := uniseg.GraphemeClusterCount(reqBody.Caption)
	if titleLength > 150 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Title is too long."}))
	}
	if captionLength > 200 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Caption is too long."}))
	}

	// Update the fields
	var base = models.Base{Id: c.Params("postId")}
	var post = models.Post{Base: base, ProfileId: reqProfile.Id}
	if err := configs.Database.Model(&post).Updates(map[string]interface{}{"title": reqBody.Title, "caption": reqBody.Caption, "is_archived": *reqBody.IsArchived}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been successfully updated."}))
}

func DeletePost(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var post models.Post
	if err := configs.Database.Model(&models.Post{}).Delete(&post, "id = ? AND profile_id = ?", c.Params("postId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "The post was deleted successfully."}))
}
