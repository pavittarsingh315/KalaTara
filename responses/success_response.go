package responses

import (
	"github.com/gofiber/fiber/v2"
)

type successResponse struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    *fiber.Map `json:"data"`
}

func NewSuccessResponse(status int, data *fiber.Map) successResponse {
	response := successResponse{}
	response.Message = "Success"
	response.Status = status
	response.Data = data
	return response
}
