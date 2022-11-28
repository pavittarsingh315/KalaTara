package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Handles the query parameters: page and limit. Then it adds the page, limit, and offset to c.Locals(...)
func PaginationHandler(c *fiber.Ctx) error {
	var page, limit int = 1, 10

	if pageParam := c.Query("page"); pageParam != "" {
		pageNum, _ := strconv.Atoi(pageParam)
		if pageNum >= 1 {
			page = pageNum
		}
	}

	if limitParam := c.Query("limit"); limitParam != "" {
		limitValue, _ := strconv.Atoi(limitParam)
		if limitValue >= 1 {
			if limitValue > 25 {
				limit = 25
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
