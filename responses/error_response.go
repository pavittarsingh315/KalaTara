package responses

import (
	"github.com/gofiber/fiber/v2"
)

type errorResponse struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    *fiber.Map `json:"data"`
}

func NewErrorResponse(status int, data *fiber.Map) errorResponse {
	response := errorResponse{}
	response.Message = "Error"
	response.Status = status
	response.Data = data
	return response
}
