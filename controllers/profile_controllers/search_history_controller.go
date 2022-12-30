package profilecontrollers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/configs"
	"nerajima.com/NeraJima/models"
	"nerajima.com/NeraJima/responses"
)

func AddToSearchHistory(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)
	reqBody := struct {
		Query string `json:"query"`
	}{}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Bad request..."}))
	}

	// Check if all fields are included
	if reqBody.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.NewErrorResponse(fiber.StatusBadRequest, &fiber.Map{"data": "Please include all fields."}))
	}

	reqBody.Query = strings.TrimSpace(reqBody.Query) // remove leading and trailing whitespace

	// Get length of history
	historyLength := configs.Database.Model(&reqProfile).Association("SearchHistory").Count()
	if historyLength > 20 { // delete bottom 8
		if err := configs.Database.Model(&models.SearchHistory{}).Limit(8).Order("created_at ASC").Delete(&models.SearchHistory{}, "profile_id = ?", reqProfile.Id).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
		}
	}

	newHistoryObj := models.SearchHistory{
		ProfileId: reqProfile.Id,
		Query:     reqBody.Query,
	}
	if err := configs.Database.Table("search_histories").Create(&newHistoryObj).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Search added to history"}))
}

func RemoveFromSearchHistory(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Delete the object
	var historyObj models.SearchHistory
	if err := configs.Database.Table("search_histories").Delete(&historyObj, "profile_id = ? AND id = ?", reqProfile.Id, c.Params("searchHistoryId")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Search removed from history."}))
}

func ClearSearchHistory(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Clear associated SearchHistory objects
	var historyObj models.SearchHistory
	if err := configs.Database.Table("search_histories").Delete(&historyObj, "profile_id = ?", reqProfile.Id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": "Search history cleared."}))
}

func GetSearchHistory(c *fiber.Ctx) error {
	var reqProfile models.Profile = c.Locals("profile").(models.Profile)

	// Find associated SearchHistory objects
	var history []models.SearchHistory
	if err := configs.Database.Model(&reqProfile).Order("search_histories.created_at DESC").Association("SearchHistory").Find(&history); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.NewErrorResponse(fiber.StatusInternalServerError, &fiber.Map{"data": "Unexpected Error. Please try again."}))
	}

	return c.Status(fiber.StatusOK).JSON(responses.NewSuccessResponse(fiber.StatusOK, &fiber.Map{"data": history}))
}
