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
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot invite yourself."}, nil))
	}

	// More info on this query: https://gorm.io/docs/advanced_query.html#FirstOrCreate

	newSubscriberObj := models.ProfileSubscriber{
		ProfileId:    reqProfile.Id,
		SubscriberId: c.Params("profileId"),
	}
	newSubscriberObjAttributes := models.ProfileSubscriber{
		IsInvite:  true,
		IsRequest: false,
	}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Where(newSubscriberObj).Attrs(newSubscriberObjAttributes).FirstOrCreate(&newSubscriberObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been sent."}))
}

func CancelInviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var subscriber models.ProfileSubscriber
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Delete(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_invite = ? AND is_accepted = ?", reqProfile.Id, c.Params("profileId"), true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been canceled."}))
}

func AcceptInviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Where("profile_id = ? AND subscriber_id = ? AND is_invite = ? AND is_accepted = ?", c.Params("senderId"), reqProfile.Id, true, false).Update("is_accepted", "1").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been accepted."}))
}

func DeclineInviteToSubscribersList(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var subscriber models.ProfileSubscriber
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Delete(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_invite = ? AND is_accepted = ?", c.Params("senderId"), reqProfile.Id, true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Invite has been declined."}))
}

func RequestToSubscribe(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot request yourself."}, nil))
	}

	// More info on this query: https://gorm.io/docs/advanced_query.html#FirstOrCreate

	newSubscriberObj := models.ProfileSubscriber{
		ProfileId:    c.Params("profileId"),
		SubscriberId: reqProfile.Id,
	}
	newSubscriberObjAttributes := models.ProfileSubscriber{
		IsInvite:  false,
		IsRequest: true,
	}
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Where(newSubscriberObj).Attrs(newSubscriberObjAttributes).FirstOrCreate(&newSubscriberObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Request has been sent."}))
}

func CancelRequestToSubscribe(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var subscriber models.ProfileSubscriber
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Delete(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_request = ? AND is_accepted = ?", c.Params("profileId"), reqProfile.Id, true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Request has been canceled."}))
}

func AcceptRequestToSubscribe(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Where("profile_id = ? AND subscriber_id = ? AND is_request = ? AND is_accepted = ?", reqProfile.Id, c.Params("senderId"), true, false).Update("is_accepted", "1").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Request has been accepted."}))
}

func DeclineRequestToSubscribe(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var subscriber models.ProfileSubscriber
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Delete(&subscriber, "profile_id = ? AND subscriber_id = ? AND is_request = ? AND is_accepted = ?", reqProfile.Id, c.Params("senderId"), true, false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Request has been declined."}))
}

func RemoveASubscriber(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot remove yourself."}, nil))
	}

	// Delete the object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var subscriberObj models.ProfileSubscriber
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Delete(&subscriberObj, "profile_id = ? AND subscriber_id = ?", reqProfile.Id, c.Params("profileId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Subscriber has been removed."}))
}

func UnsubscribeFromUser(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	if reqProfile.Id == c.Params("profileId") {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "You cannot unsubscribe from yourself."}, nil))
	}

	// Delete the object
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var subscriberObj models.ProfileSubscriber
	if err := configs.Database.WithContext(dbCtx).Table("profile_subscribers").Delete(&subscriberObj, "profile_id = ? AND subscriber_id = ?", c.Params("profileId"), reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Subscription has been canceled."}))
}

func GetSubscribers(c *fiber.Ctx) error {
	var page int = c.Locals("page").(int)
	var limit int = c.Locals("limit").(int)
	var offset int = c.Locals("offset").(int)
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Get subscribers(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	var subscribers = []models.MiniProfile{}
	if err := configs.Database.WithContext(dbCtx).Model(&reqProfile).Offset(offset).Limit(limit).Order("profile_subscribers.created_at DESC").Where("is_accepted = ?", true).Where("username LIKE ?", regexMatch).Association("Subscribers").Find(&subscribers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of subscribers
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	numSubscribers := configs.Database.WithContext(dbCtx2).Model(&reqProfile).Where("is_accepted = ?", true).Where("username LIKE ?", regexMatch).Association("Subscribers").Count()

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
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

	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := configs.Database.Table("profiles").
		Joins("JOIN profile_subscribers ON profile_subscribers.subscriber_id = ? AND profile_subscribers.profile_id = profiles.id", reqProfile.Id).
		Where("is_accepted = ?", true).
		Where("username LIKE ?", regexMatch)

	// Get subscriptions(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var subscriptions = []models.MiniProfile{}
	if err := query.WithContext(dbCtx).Order("profile_subscribers.created_at DESC").Limit(limit).Offset(offset).Scan(&subscriptions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of subscriptions
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numSubscriptions int64
	if err := query.WithContext(dbCtx2).Count(&numSubscriptions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
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

	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := configs.Database.Table("profiles").
		Joins("JOIN profile_subscribers ON profile_subscribers.subscriber_id = profiles.id AND profile_subscribers.profile_id = ?", reqProfile.Id).
		Where("is_accepted = ? AND is_invite = ? AND username LIKE ?", false, true, regexMatch)

	// Get invites sent(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var invitesSent = []models.MiniProfile{}
	if err := query.WithContext(dbCtx).Order("profile_subscribers.created_at DESC").Limit(limit).Offset(offset).Scan(&invitesSent).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of invites sent
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numInvitesSent int64
	if err := query.WithContext(dbCtx2).Count(&numInvitesSent).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
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

	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := configs.Database.Table("profiles").
		Joins("JOIN profile_subscribers ON profile_subscribers.subscriber_id = ? AND profile_subscribers.profile_id = profiles.id", reqProfile.Id).
		Where("is_accepted = ? AND is_invite = ? AND username LIKE ?", false, true, regexMatch)

	// Get invites sent(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var invitesReceived = []models.MiniProfile{}
	if err := query.WithContext(dbCtx).Order("profile_subscribers.created_at DESC").Limit(limit).Offset(offset).Scan(&invitesReceived).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of invites received
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numInvitesReceived int64
	if err := query.WithContext(dbCtx2).Count(&numInvitesReceived).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
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

	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := configs.Database.Table("profiles").
		Joins("JOIN profile_subscribers ON profile_subscribers.subscriber_id = ? AND profile_subscribers.profile_id = profiles.id", reqProfile.Id).
		Where("is_accepted = ? AND is_request = ? AND username LIKE ?", false, true, regexMatch)

	// Get requests sent(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var requestsSent = []models.MiniProfile{}
	if err := query.WithContext(dbCtx).Order("profile_subscribers.created_at DESC").Limit(limit).Offset(offset).Scan(&requestsSent).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of requests sent
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numRequestsSent int64
	if err := query.WithContext(dbCtx2).Count(&numRequestsSent).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
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

	regexMatch := fmt.Sprintf("%s%%", c.Query("filter")) // for more information on regex matching in sql, visit https://www.freecodecamp.org/news/sql-contains-string-sql-regex-example-query/
	query := configs.Database.Table("profiles").
		Joins("JOIN profile_subscribers ON profile_subscribers.profile_id = ? AND profile_subscribers.subscriber_id = profiles.id", reqProfile.Id).
		Where("is_accepted = ? AND is_request = ? AND username LIKE ?", false, true, regexMatch)

	// Get requests sent(paginated)
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var requestsReceived = []models.MiniProfile{}
	if err := query.WithContext(dbCtx).Order("profile_subscribers.created_at DESC").Limit(limit).Offset(offset).Scan(&requestsReceived).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Get total number of requests sent
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	var numRequestsReceived int64
	if err := query.WithContext(dbCtx2).Count(&numRequestsReceived).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{
		"data": &fiber.Map{
			"current_page": page,
			"per_page":     limit,
			"last_page":    int(math.Ceil(float64(numRequestsReceived) / float64(limit))),
			"data":         requestsReceived,
		},
	}))
}
