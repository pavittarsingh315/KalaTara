package routes

import (
	"github.com/gofiber/fiber/v2"
	authcontrollers "nerajima.com/NeraJima/controllers/auth_controllers"
)

func AuthRouter(group fiber.Router) {
	router := group.Group("/auth")

	router.Post("/register/initiate", authcontrollers.InitiateRegistration)
	router.Post("/register/finalize", authcontrollers.FinalizeRegistration)
}
