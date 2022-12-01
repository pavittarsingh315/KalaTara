package profilecontrollers

import (
	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func InviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot invite yourself."}))
	}

	// check if user being invited exists
	var toBeInvitedProfile models.Profile
	if err := configs.Database.Model(&models.Profile{}).Find(&toBeInvitedProfile, "id = ?", c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
	}
	if toBeInvitedProfile.Id == "" { // Id field is empty => user does not exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "The user you are trying to invite does not exist."}))
	}

	// check if user is already a sub OR (if invite is already sent OR user requested to be a subscriber)
	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriber, "profile_id = ? AND subscriber_id = ?", reqProfile.Id, c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
	}
	if subscriber.ProfileId != "" && subscriber.SubscriberId != "" { // if both fields are populated => reqUser is eith already subscribed or an invite/request exists
		if subscriber.IsAccepted {
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This user is already one of your subscribers."}))
		} else {
			if subscriber.IsInvite {
				return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You have already invited this user."}))
			} else if subscriber.IsRequest {
				return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This user has already requested to be one of your subscribers. Accept the request."}))
			}
		}
	}

	newSubscriberObj := models.ProfileSubscriber{
		ProfileId:    reqProfile.Id,
		SubscriberId: toBeInvitedProfile.Id,
		IsInvite:     true,
		IsRequest:    false,
	}
	if err := configs.Database.Table("profile_subscribers").Create(&newSubscriberObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been sent."}))
}

func CancelInviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Delete(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_invite = ? AND is_accepted = ?", reqProfile.Id, c.Params("profileId"), true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been canceled."}))
}

func AcceptInviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_invite = ? AND is_accepted = ?", c.Params("senderId"), reqProfile.Id, true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
	}
	if subscriber.ProfileId == "" && subscriber.SubscriberId == "" { // if both fields are empty, there doesn't exist such an invite
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "This invite does not exist."}))
	}

	if err := configs.Database.Model(&subscriber).Update("is_accepted", "1").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been accepted."}))
}

func DeclineInviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_invite = ? AND is_accepted = ?", c.Params("senderId"), reqProfile.Id, true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
	}
	if subscriber.ProfileId == "" && subscriber.SubscriberId == "" { // if both fields are empty, there doesn't exist such an invite
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "This invite does not exist."}))
	}

	if err := configs.Database.Delete(&subscriber).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been declined."}))
}
