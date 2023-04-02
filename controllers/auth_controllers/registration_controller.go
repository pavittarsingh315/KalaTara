package authcontrollers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rivo/uniseg"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/configs/cache"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
	"nerajima.com/NeraJima/utils"
)

func InitiateRegistration(c *fiber.Ctx) error {
	reqBody := struct {
		Contact  string    `json:"contact"`
		Username string    `json:"username"`
		Name     string    `json:"name"`
		Password string    `json:"password"`
		Birthday time.Time `json:"birthday"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}, err))
	}

	// Check if all fields are included
	if reqBody.Contact == "" || reqBody.Username == "" || reqBody.Name == "" || reqBody.Password == "" || reqBody.Birthday.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}, nil))
	}

	reqBody.Name = strings.TrimSpace(reqBody.Name)                                    // remove leading and trailing whitespace
	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", ""))   // remove all whitespace and make lowercase
	reqBody.Username = strings.ToLower(strings.ReplaceAll(reqBody.Username, " ", "")) // remove all whitespace and make lowercase

	// Validate request body lengths
	usernameLength := uniseg.GraphemeClusterCount(reqBody.Username)
	nameLength := uniseg.GraphemeClusterCount(reqBody.Name)
	contactLength := uniseg.GraphemeClusterCount(reqBody.Contact)
	passwordLength := uniseg.GraphemeClusterCount(reqBody.Password)
	if usernameLength < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username is too short."}, nil))
	}
	if usernameLength > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username is too long."}, nil))
	}
	if nameLength > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Name is too long."}, nil))
	}
	if contactLength > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Contact is too long."}, nil))
	}
	if passwordLength < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Password is too short."}, nil))
	}
	if length := len(reqBody.Password); length > 64 { // Since the max length password supported by bcrypt is 72 bytes, we check the length of the string in bytes. I made max length 64 to be safe rather than 72.
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Password too long."}, nil))
	}

	// Check if username is taken
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var profile models.Profile
	if err := configs.Database.WithContext(dbCtx).Model(&models.Profile{}).Find(&profile, "username = ?", reqBody.Username).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	if profile.Username != "" { // username field is not empty => profile with username exists
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username is taken."}, nil))
	}

	// Check if contact is taken
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	contactIsEmail := utils.ValidateEmail(reqBody.Contact)
	var user models.User
	if err := configs.Database.WithContext(dbCtx2).Model(&models.User{}).Find(&user, "contact = ?", reqBody.Contact).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	if user.Contact != "" { // contact field is not empty => user with contact exists
		errorMsg := "Contact already in use."
		if contactIsEmail {
			errorMsg = "Email address already in use."
		}
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": errorMsg}, nil))
	}

	// Check if registration is already initiated
	cacheCtx, cacheCancel := cache.NewCacheContext()
	defer cacheCancel()
	var key = cache.NewUserConfirmCodeKey(reqBody.Contact)
	var confirmationCode string
	if err := cache.Get(cacheCtx, key, &confirmationCode); err == nil { // no error => key exists ie hasnt expired
		cacheCtx, cacheCancel := cache.NewCacheContext()
		defer cacheCancel()
		dur, _ := cache.ExpiresIn(cacheCtx, key)
		message := fmt.Sprintf("Try again in %s.", utils.SecondsToString(int64(dur.Seconds())))
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": message}, nil))
	} else if err != redis.Nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Create new user confirm code in cache
	cacheCtx2, cacheCancel2 := cache.NewCacheContext()
	defer cacheCancel2()
	var code, err = utils.GenerateRandomCode(6)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	var hash, err2 = utils.HashPassword(code)
	if err2 != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err2))
	}
	var exp = cache.NewUserConfirmCodeExp
	if err := cache.Set(cacheCtx2, key, hash, exp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	if contactIsEmail {
		go utils.SendRegistrationEmail(reqBody.Name, reqBody.Contact, code)
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "An email has been sent with a verification code."}))
	} else {
		go utils.SendRegistrationText(code, reqBody.Contact)
		return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "A text has been sent with a verification code."}))
	}
}

func FinalizeRegistration(c *fiber.Ctx) error {
	reqBody := struct {
		Code     string    `json:"code"`
		Contact  string    `json:"contact"`
		Username string    `json:"username"`
		Name     string    `json:"name"`
		Password string    `json:"password"`
		Birthday time.Time `json:"birthday"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}, err))
	}

	// Check if all fields are included
	if reqBody.Code == "" || reqBody.Contact == "" || reqBody.Username == "" || reqBody.Name == "" || reqBody.Password == "" || reqBody.Birthday.IsZero() {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}, nil))
	}

	reqBody.Name = strings.TrimSpace(reqBody.Name)                                    // remove leading and trailing whitespace
	reqBody.Contact = strings.ToLower(strings.ReplaceAll(reqBody.Contact, " ", ""))   // remove all whitespace and make lowercase
	reqBody.Username = strings.ToLower(strings.ReplaceAll(reqBody.Username, " ", "")) // remove all whitespace and make lowercase

	// Validate request body lengths
	usernameLength := uniseg.GraphemeClusterCount(reqBody.Username)
	nameLength := uniseg.GraphemeClusterCount(reqBody.Name)
	contactLength := uniseg.GraphemeClusterCount(reqBody.Contact)
	passwordLength := uniseg.GraphemeClusterCount(reqBody.Password)
	if usernameLength < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username too short."}, nil))
	}
	if usernameLength > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username too long."}, nil))
	}
	if nameLength > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Name too long."}, nil))
	}
	if contactLength > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Contact too long."}, nil))
	}
	if passwordLength < 10 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Password too short."}, nil))
	}
	if length := len(reqBody.Password); length > 64 { // Since the max length password supported by bcrypt is 72 bytes, we check the length of the string in bytes. I made max length 64 to be safe rather than 72.
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Password too long."}, nil))
	}

	// Check if username is taken
	dbCtx, dbCancel := configs.NewQueryContext()
	defer dbCancel()
	var profile models.Profile
	if err := configs.Database.WithContext(dbCtx).Model(&models.Profile{}).Find(&profile, "username = ?", reqBody.Username).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}
	if profile.Username != "" { // username field is not empty => profile with username exists
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Username is taken."}, nil))
	}

	// Get confirmation code
	cacheCtx, cacheCancel := cache.NewCacheContext()
	defer cacheCancel()
	var key = cache.NewUserConfirmCodeKey(reqBody.Contact)
	var confirmationCode string
	if err := cache.Get(cacheCtx, key, &confirmationCode); err != nil {
		if err == redis.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Code has expired. Please restart the registration process."}, nil))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
		}
	}

	// Check if user provided code is correct
	if !utils.VerifyPassword(confirmationCode, reqBody.Code) {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Incorrect Code."}, nil))
	}

	var hash, err = utils.HashPassword(reqBody.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Create User
	newUser := models.User{
		Name:      reqBody.Name,
		Contact:   reqBody.Contact,
		Password:  hash,
		Role:      "user",
		Strikes:   0,
		Birthday:  reqBody.Birthday,
		LastLogin: time.Now(),
		BanTill:   time.Now(),
		Profile: models.Profile{
			Username:   reqBody.Username,
			Name:       reqBody.Name,
			Bio:        "🚀🚀🚀🚀🚀🚀🚀🚀",
			Avatar:     "https://nerajima.s3.us-west-1.amazonaws.com/default.jpg",
			MiniAvatar: "https://nerajima.s3.us-west-1.amazonaws.com/default.jpg",
			Birthday:   reqBody.Birthday,
		},
	}
	dbCtx2, dbCancel2 := configs.NewQueryContext()
	defer dbCancel2()
	if err := configs.Database.WithContext(dbCtx2).Model(&models.User{}).Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	// Delete confirmation code from cache
	cacheCtx2, cacheCancel2 := cache.NewCacheContext()
	defer cacheCancel2()
	cache.Delete(cacheCtx2, key)

	// Generate auth tokens
	access, refresh := utils.GenAuthTokens(newUser.Id)

	// Cache profile
	cacheCtx3, cacheCancel3 := cache.NewCacheContext()
	defer cacheCancel3()
	key = cache.ProfileKey(newUser.Id)
	var exp = cache.ProfileExp
	if err := cache.Set(cacheCtx3, key, newUser.Profile, exp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}, err))
	}

	return c.Status(fiber.StatusOK).JSON(
		responses.NewSuccessResponse(
			fiber.StatusOK,
			&fiber.Map{
				"data": &fiber.Map{
					"access":  access,
					"refresh": refresh,
					"profile": newUser.Profile,
				},
			},
		),
	)
}
