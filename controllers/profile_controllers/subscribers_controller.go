package profilecontrollers

import (
	"fmt"
	"math"

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
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if toBeInvitedProfile.Id == "" { // Id field is empty => user does not exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "The user you are trying to invite does not exist."}))
	}

	// check if user is already a sub OR (if invite is already sent OR user requested to be a subscriber)
	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriber, "profile_id = ? AND subscriber_id = ?", reqProfile.Id, c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if subscriber.ProfileId != "" && subscriber.SubscriberId != "" { // if both fields are populated => user is either already subscribed or an invite/request exists
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
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been sent."}))
}

func CancelInviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Delete(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_invite = ? AND is_accepted = ?", reqProfile.Id, c.Params("profileId"), true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been canceled."}))
}

func AcceptInviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_invite = ? AND is_accepted = ?", c.Params("senderId"), reqProfile.Id, true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if subscriber.ProfileId == "" && subscriber.SubscriberId == "" { // if both fields are empty, there doesn't exist such an invite
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "This invite does not exist."}))
	}

	if err := configs.Database.Model(&subscriber).Update("is_accepted", "1").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been accepted."}))
}

func DeclineInviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_invite = ? AND is_accepted = ?", c.Params("senderId"), reqProfile.Id, true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if subscriber.ProfileId == "" && subscriber.SubscriberId == "" { // if both fields are empty, there doesn't exist such an invite
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "This invite does not exist."}))
	}

	if err := configs.Database.Delete(&subscriber).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been declined."}))
}

func RequestToSubscribe(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot request yourself."}))
	}

	// check if user being requested exists
	var toBeRequestedProfile models.Profile
	if err := configs.Database.Model(&models.Profile{}).Find(&toBeRequestedProfile, "id = ?", c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if toBeRequestedProfile.Id == "" { // Id field is empty => user does not exist
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "The user you are trying to request does not exist."}))
	}

	// check if user is already a sub OR (if request is already sent OR user being requested already sent an invite)
	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriber, "profile_id = ? AND subscriber_id = ?", c.Params("profileId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if subscriber.ProfileId != "" && subscriber.SubscriberId != "" { // if both fields are populated => reqUser is either already subscribed or an invite/request exists
		if subscriber.IsAccepted {
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You are already subscribed to this user."}))
		} else {
			if subscriber.IsInvite {
				return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "This user has sent you an invite to subscribe to them. Accept the request"}))
			} else if subscriber.IsRequest {
				return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You've already requested to subscribe to this user."}))
			}
		}
	}

	newSubscriberObj := models.ProfileSubscriber{
		ProfileId:    toBeRequestedProfile.Id,
		SubscriberId: reqProfile.Id,
		IsInvite:     false,
		IsRequest:    true,
	}
	if err := configs.Database.Table("profile_subscribers").Create(&newSubscriberObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Request has been sent."}))
}

func CancelRequestToSubscribe(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Delete(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_request = ? AND is_accepted = ?", c.Params("profileId"), reqProfile.Id, true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Request has been canceled."}))
}

func AcceptRequestToSubscribe(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_request = ? AND is_accepted = ?", reqProfile.Id, c.Params("senderId"), true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if subscriber.ProfileId == "" && subscriber.SubscriberId == "" { // if both fields are empty, there doesn't exist such a request
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "This request does not exist."}))
	}

	if err := configs.Database.Model(&subscriber).Update("is_accepted", "1").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Request has been accepted."}))
}

func DeclineRequestToSubscribe(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	var subscriber models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Find(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_request = ? AND is_accepted = ?", reqProfile.Id, c.Params("senderId"), true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}
	if subscriber.ProfileId == "" && subscriber.SubscriberId == "" { // if both fields are empty, there doesn't exist such a request
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "This request does not exist."}))
	}

	if err := configs.Database.Delete(&subscriber).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Request has been declined."}))
}

func RemoveASubscriber(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot remove yourself."}))
	}

	// Delete the object
	var subscriberObj models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Delete(&subscriberObj, "profile_id = ? AND subscriber_id = ?", reqProfile.Id, c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Subscriber has been removed."}))
}

func UnsubscribeFromUser(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot unsubscribe from yourself."}))
	}

	// Delete the object
	var subscriberObj models.ProfileSubscriber
	if err := configs.Database.Table("profile_subscribers").Delete(&subscriberObj, "profile_id = ? AND subscriber_id = ?", c.Params("profileId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Subscription has been canceled."}))
}

func GetSubscribers(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Get subscribers(paginated)
	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	var subscribers = []models.MiniProfile{}
	if err := configs.Database.Model(&reqProfile).Offset(offset).Limit(limit).Order("profile_subscribers.created_at DESC").Where("is_accepted = ?", true).Where("username LIKE ? OR name LIKE ?", regexMatch, regexMatch).Association("Subscribers").Find(&subscribers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of subscribers
	numSubscribers := configs.Database.Model(&reqProfile).Where("is_accepted = ?", true).Where("username LIKE ? OR name LIKE ?", regexMatch, regexMatch).Association("Subscribers").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numSubscribers) / float64(limit))),
			"data":         subscribers,
		},
	}))
}

func GetSubscriptions(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	/*
	   IMPORTANT:
	      To build the raw queries below, I used the default gorm logger in Info mode to inspect the queries made by the GetSubscribers endpoint when gettings the subscribers list and numsubscribers.
	      I then used those queries and altered them to build the queries seen below.
	*/

	// Get subscriptions(paginated)
	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := fmt.Sprintf("SELECT profiles.id, profiles.created_at, profiles.updated_at, profiles.user_id, profiles.username, profiles.name, profiles.bio, profiles.avatar, profiles.mini_avatar, profiles.birthday FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profiles.id WHERE is_accepted = %d AND (username LIKE \"%s\" OR name LIKE \"%s\") ORDER BY profile_subscribers.created_at DESC LIMIT %d OFFSET %d", reqProfile.Id, 1, regexMatch, regexMatch, limit, offset)
	var subscriptions = []models.MiniProfile{}
	if err := configs.Database.Raw(query).Scan(&subscriptions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of subscriptions
	var numSubscriptions int
	query2 := fmt.Sprintf("SELECT count(*) FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profiles.id WHERE is_accepted = %d AND (username LIKE \"%s\" OR name LIKE \"%s\")", reqProfile.Id, 1, regexMatch, regexMatch)
	if err := configs.Database.Raw(query2).Scan(&numSubscriptions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numSubscriptions) / float64(limit))),
			"data":         subscriptions,
		},
	}))
}

func GetInvitesSent(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	/*
	   IMPORTANT:
	      To build the raw queries below, I used the raw query from the GetSubscriptions endpoint made a few changes
	*/

	// Get invites sent(paginated)
	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := fmt.Sprintf("SELECT profiles.id, profiles.created_at, profiles.updated_at, profiles.user_id, profiles.username, profiles.name, profiles.bio, profiles.avatar, profiles.mini_avatar, profiles.birthday FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = profiles.id AND profile_subscribers.profile_id = \"%s\" WHERE is_accepted = %d AND is_invite = %d AND (username LIKE \"%s\" OR name LIKE \"%s\") ORDER BY profile_subscribers.created_at DESC LIMIT %d OFFSET %d", reqProfile.Id, 0, 1, regexMatch, regexMatch, limit, offset)
	var invitesSent = []models.MiniProfile{}
	if err := configs.Database.Raw(query).Scan(&invitesSent).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of invites sent
	// SELECT count(*) FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profiles.id WHERE is_accepted = %d AND (username LIKE \"%s\" OR name LIKE \"%s\")
	var numInvitesSent int
	query2 := fmt.Sprintf("SELECT count(*) FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = profiles.id AND profile_subscribers.profile_id = \"%s\" WHERE is_accepted = %d AND is_invite = %d AND (username LIKE \"%s\" OR name LIKE \"%s\")", reqProfile.Id, 0, 1, regexMatch, regexMatch)
	if err := configs.Database.Raw(query2).Scan(&numInvitesSent).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numInvitesSent) / float64(limit))),
			"data":         invitesSent,
		},
	}))
}

func GetInvitesReceived(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	/*
	   IMPORTANT:
	      To build the raw queries below, I used the raw query from the GetSubscriptions endpoint made a few changes
	*/

	// Get invites sent(paginated)
	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := fmt.Sprintf("SELECT profiles.id, profiles.created_at, profiles.updated_at, profiles.user_id, profiles.username, profiles.name, profiles.bio, profiles.avatar, profiles.mini_avatar, profiles.birthday FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profiles.id WHERE is_accepted = %d AND is_invite = %d AND (username LIKE \"%s\" OR name LIKE \"%s\") ORDER BY profile_subscribers.created_at DESC LIMIT %d OFFSET %d", reqProfile.Id, 0, 1, regexMatch, regexMatch, limit, offset)
	var invitesReceived = []models.MiniProfile{}
	if err := configs.Database.Raw(query).Scan(&invitesReceived).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of invites received
	var numInvitesReceived int
	query2 := fmt.Sprintf("SELECT count(*) FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profiles.id WHERE is_accepted = %d AND is_invite = %d AND (username LIKE \"%s\" OR name LIKE \"%s\")", reqProfile.Id, 0, 1, regexMatch, regexMatch)
	if err := configs.Database.Raw(query2).Scan(&numInvitesReceived).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numInvitesReceived) / float64(limit))),
			"data":         invitesReceived,
		},
	}))
}

func GetRequestsSent(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	/*
	   IMPORTANT:
	      To build the raw queries below, I used the raw query from the GetSubscriptions endpoint made a few changes
	*/

	// Get requests sent(paginated)
	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := fmt.Sprintf("SELECT profiles.id, profiles.created_at, profiles.updated_at, profiles.user_id, profiles.username, profiles.name, profiles.bio, profiles.avatar, profiles.mini_avatar, profiles.birthday FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profiles.id WHERE is_accepted = %d AND is_request = %d AND (username LIKE \"%s\" OR name LIKE \"%s\") ORDER BY profile_subscribers.created_at DESC LIMIT %d OFFSET %d", reqProfile.Id, 0, 1, regexMatch, regexMatch, limit, offset)
	var requestsSent = []models.MiniProfile{}
	if err := configs.Database.Raw(query).Scan(&requestsSent).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of requests sent
	var numRequestsSent int
	query2 := fmt.Sprintf("SELECT count(*) FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = \"%s\" AND profile_subscribers.profile_id = profiles.id WHERE is_accepted = %d AND is_request = %d AND (username LIKE \"%s\" OR name LIKE \"%s\")", reqProfile.Id, 0, 1, regexMatch, regexMatch)
	if err := configs.Database.Raw(query2).Scan(&numRequestsSent).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numRequestsSent) / float64(limit))),
			"data":         requestsSent,
		},
	}))
}

func GetRequestsReceived(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	/*
	   IMPORTANT:
	      To build the raw queries below, I used the raw query from the GetSubscriptions endpoint made a few changes
	*/

	// Get requests sent(paginated)
	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := fmt.Sprintf("SELECT profiles.id, profiles.created_at, profiles.updated_at, profiles.user_id, profiles.username, profiles.name, profiles.bio, profiles.avatar, profiles.mini_avatar, profiles.birthday FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = profiles.id AND profile_subscribers.profile_id = \"%s\" WHERE is_accepted = %d AND is_request = %d AND (username LIKE \"%s\" OR name LIKE \"%s\") ORDER BY profile_subscribers.created_at DESC LIMIT %d OFFSET %d", reqProfile.Id, 0, 1, regexMatch, regexMatch, limit, offset)
	var requestsReceived = []models.MiniProfile{}
	if err := configs.Database.Raw(query).Scan(&requestsReceived).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	// Get total number of requests sent
	var numRequestsReceived int
	query2 := fmt.Sprintf("SELECT count(*) FROM profiles JOIN profile_subscribers ON profile_subscribers.subscriber_id = profiles.id AND profile_subscribers.profile_id = \"%s\" WHERE is_accepted = %d AND is_request = %d AND (username LIKE \"%s\" OR name LIKE \"%s\")", reqProfile.Id, 0, 1, regexMatch, regexMatch)
	if err := configs.Database.Raw(query2).Scan(&numRequestsReceived).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"last_page":    int(math.Ceil(float64(numRequestsReceived) / float64(limit))),
			"data":         requestsReceived,
		},
	}))
}
