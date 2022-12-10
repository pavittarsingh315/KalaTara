package routes

import (
	"github.com/gofiber/fiber/v2"
	profilecontrollers "nerajima.com/NeraJima/controllers/profile_controllers"
	"nerajima.com/NeraJima/middleware"
)

func ProfileRouter(group fiber.Router) {
	router := group.Group("/profile") // domain/api/profile

	editRouter(router)
	followersRouter(router)
	searchHistoryRouter(router)
	subscribersRouter(router)
}

func editRouter(group fiber.Router) {
	router := group.Group("/edit") // domain/api/profile/edit

	router.Put("/username", middleware.UserAuthHandler, profilecontrollers.EditUsername)
	router.Put("/name", middleware.UserAuthHandler, profilecontrollers.EditName)
	router.Put("/bio", middleware.UserAuthHandler, profilecontrollers.EditBio)
	router.Put("/avatar", middleware.UserAuthHandler, profilecontrollers.EditAvatar)
}

func followersRouter(group fiber.Router) {
	router := group.Group("/followers") // domain/api/profile/followers

	router.Post("/follow/:profileId", middleware.UserAuthHandler, profilecontrollers.FollowAUser)
	router.Delete("/unfollow/:profileId", middleware.UserAuthHandler, profilecontrollers.UnfollowAUser)
	router.Delete("/remove/:profileId", middleware.UserAuthHandler, profilecontrollers.RemoveAFollower)
	router.Get("/get-followers/:profileId", middleware.UserAuthHandler, middleware.PaginationHandler, profilecontrollers.GetFollowers)
	router.Get("/get-following/:profileId", middleware.UserAuthHandler, middleware.PaginationHandler, profilecontrollers.GetFollowing)
}

func searchHistoryRouter(group fiber.Router) {
	router := group.Group("/search-history") // domain/api/profile/search-history

	router.Post("/add", middleware.UserAuthHandler, profilecontrollers.AddToSearchHistory)
	router.Delete("/remove/:searchHistoryId", middleware.UserAuthHandler, profilecontrollers.RemoveFromSearchHistory)
	router.Delete("/clear", middleware.UserAuthHandler, profilecontrollers.ClearSearchHistory)
	router.Get("/get", middleware.UserAuthHandler, profilecontrollers.GetSearchHistory)
}

func subscribersRouter(group fiber.Router) {
	router := group.Group("/subscribers") // domain/api/profile/subscribers

	router.Post("/invite/:profileId", middleware.UserAuthHandler, profilecontrollers.InviteToSubscribersList)
	router.Delete("/invite/cancel/:profileId", middleware.UserAuthHandler, profilecontrollers.CancelInviteToSubscribersList)
	router.Put("/invite/accept/:senderId", middleware.UserAuthHandler, profilecontrollers.AcceptInviteToSubscribersList)
	router.Delete("/invite/decline/:senderId", middleware.UserAuthHandler, profilecontrollers.DeclineInviteToSubscribersList)

	router.Post("/request/:profileId", middleware.UserAuthHandler, profilecontrollers.RequestToSubscribe)
	router.Delete("/request/cancel/:profileId", middleware.UserAuthHandler, profilecontrollers.CancelRequestToSubscribe)
	router.Put("/request/accept/:senderId", middleware.UserAuthHandler, profilecontrollers.AcceptRequestToSubscribe)
	router.Delete("/request/decline/:senderId", middleware.UserAuthHandler, profilecontrollers.DeclineRequestToSubscribe)

	router.Delete("/remove/:profileId", middleware.UserAuthHandler, profilecontrollers.RemoveASubscriber)
	router.Delete("/unsubscribe/:profileId", middleware.UserAuthHandler, profilecontrollers.UnsubscribeFromUser)

	router.Get("/get", middleware.UserAuthHandler, middleware.PaginationHandler, profilecontrollers.GetSubscribers)
	router.Get("/subscriptions/get", middleware.UserAuthHandler, middleware.PaginationHandler, profilecontrollers.GetSubscriptions)

	router.Get("/invites/sent/get", middleware.UserAuthHandler, middleware.PaginationHandler, profilecontrollers.GetInvitesSent)
	router.Get("/invites/received/get", middleware.UserAuthHandler, middleware.PaginationHandler, profilecontrollers.GetInvitesReceived)
	router.Get("/requests/sent/get", middleware.UserAuthHandler, middleware.PaginationHandler, profilecontrollers.GetRequestsSent)
	router.Get("/requests/received/get", middleware.UserAuthHandler, middleware.PaginationHandler, profilecontrollers.GetRequestsReceived)
}
