package routes

import (
	"github.com/gofiber/fiber/v2"
	profilecontrollers "nerajima.com/NeraJima/controllers/profile_controllers"
	"nerajima.com/NeraJima/middleware"
)

func ProfileRouter(group fiber.Router) {
	router := group.Group("/profile") // domain/api/profile

	router.Put("/edit/username", middleware.UserAuthHandler, profilecontrollers.EditUsername)
	router.Put("/edit/name", middleware.UserAuthHandler, profilecontrollers.EditName)
	router.Put("/edit/bio", middleware.UserAuthHandler, profilecontrollers.EditBio)
	router.Put("/edit/avatar", middleware.UserAuthHandler, profilecontrollers.EditAvatar)

	followersRouter(router)
}

func followersRouter(group fiber.Router) {
	router := group.Group("/followers") // domain/api/profile/followers

	router.Post("/follow/:profileId", middleware.UserAuthHandler, profilecontrollers.FollowAUser)
	router.Delete("/unfollow/:profileId", middleware.UserAuthHandler, profilecontrollers.UnfollowAUser)
	router.Delete("/remove/:profileId", middleware.UserAuthHandler, profilecontrollers.RemoveAFollower)
	router.Get("/getfollowers/:profileId", middleware.UserAuthHandler, profilecontrollers.GetFollowers)
	router.Get("/getfollowing/:profileId", middleware.UserAuthHandler, profilecontrollers.GetFollowing)
}
