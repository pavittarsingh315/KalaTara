package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"nerajima.com/NeraJima/responses"
)

// Handles the query parameters: page and limit. Then it adds the page, limit, and offset to c.Locals(...)
func PaginationHandler(c *fiber.Ctx) error {
	var page, limit, maxLimit int = 1, 10, 25

	if pageParam := c.Query("page"); pageParam != "" {
		if pageNum, err := strconv.Atoi(pageParam); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(responses.NewErrorResponse(fiber.StatusUnauthorized, &fiber.Map{"data": "Could not parse page query parameter to type int."}, err))
		} else if pageNum >= 1 {
			page = pageNum
		}
	}

	if limitParam := c.Query("limit"); limitParam != "" {
		if limitValue, err := strconv.Atoi(limitParam); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(responses.NewErrorResponse(fiber.StatusUnauthorized, &fiber.Map{"data": "Could not parse limit query parameter to type int."}, err))
		} else if limitValue >= 1 {
			if limitValue > maxLimit {
				limit = maxLimit
			} else {
				limit = limitValue
			}
		}
	}

	c.Locals("page", page)
	c.Locals("limit", limit)
	c.Locals("offset", (page-1)*limit)

	return c.Next()
}
