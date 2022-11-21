package routes

import (
	"github.com/gofiber/fiber/v2"
	profilecontrollers "nerajima.com/NeraJima/controllers/profile_controllers"
	"nerajima.com/NeraJima/middleware"
)

func ProfileRouter(group fiber.Router) {
	router := group.Group("/profile")

	router.Put("/edit/username", middleware.UserAuthHandler, profilecontrollers.EditUsername)
	router.Put("/edit/name", middleware.UserAuthHandler, profilecontrollers.EditName)
	router.Put("/edit/bio", middleware.UserAuthHandler, profilecontrollers.EditBio)
	router.Put("/edit/avatar", middleware.UserAuthHandler, profilecontrollers.EditAvatar)
}
