package routes

import (
	"github.com/gofiber/fiber/v2"
	authcontrollers "nerajima.com/NeraJima/controllers/auth_controllers"
)

func AuthRouter(group fiber.Router) {
	router := group.Group("/auth") // domain/api/auth

	router.Post("/register/initiate", authcontrollers.InitiateRegistration)
	router.Post("/register/finalize", authcontrollers.FinalizeRegistration)

	router.Post("/login", authcontrollers.Login)
	router.Post("/login/token", authcontrollers.TokenLogin)

	router.Post("/password/reset/request", authcontrollers.RequestPasswordReset)
	router.Post("/password/reset/code/confirm", authcontrollers.ConfirmResetCode)
	router.Post("/password/reset/confirm", authcontrollers.ConfirmPasswordReset)
}
