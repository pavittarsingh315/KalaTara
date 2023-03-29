package responses

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

type errorResponse struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    *fiber.Map `json:"data"`
}

// Returns a new error response. If the environment is set to development, the error.Error() will be returned. Otherwise, the data will be returned.
//
// You can pass nil for the error if you want to return the data as the error response body.
func NewErrorResponse(status int, data *fiber.Map, err error) errorResponse {
	response := errorResponse{}
	response.Message = "Error"
	response.Status = status
	if err != nil && os.Getenv("APP_ENV") == "development" {
		response.Data = &fiber.Map{"data": err.Error()}
	} else {
		response.Data = data
	}
	return response
}
