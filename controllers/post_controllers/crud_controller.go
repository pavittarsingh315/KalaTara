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
	for _, mediaObj := range reqBody.Media {
		postMedia = append(postMedia, models.PostMedia{
			PostId:   newPost.Id,
			MediaUrl: mediaObj.MediaUrl,
			IsImage:  *mediaObj.IsImage,
			IsVideo:  *mediaObj.IsVideo,
			IsAudio:  *mediaObj.IsAudio,
		})
	}
	if err := configs.Database.Model(&models.PostMedia{}).Create(&postMedia).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Post has been created."}))
}

func GetPost(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Get Post"}))
}

func EditPost(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Edit Post"}))
}

func DeletePost(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Delete Post"}))
}
