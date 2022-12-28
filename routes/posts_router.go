package routes

import (
	"github.com/gofiber/fiber/v2"
	postcontrollers "nerajima.com/NeraJima/controllers/post_controllers"
	"nerajima.com/NeraJima/middleware"
)

func PostsRouter(group fiber.Router) {
	router := group.Group("/posts") // domain/api/posts

	crudRouter(router)
	specializedReadsRouter(router)
	reactionsRouter(router)
}

func crudRouter(group fiber.Router) {
	router := group // domain/api/posts

	router.Post("/create", middleware.UserAuthHandler, postcontrollers.CreatePost)
	router.Get("/get/:postId", middleware.UserAuthHandler, postcontrollers.GetPost)
	router.Put("/edit/:postId", middleware.UserAuthHandler, postcontrollers.EditPost)
	router.Delete("/delete/:postId", middleware.UserAuthHandler, postcontrollers.DeletePost)
}

func specializedReadsRouter(group fiber.Router) {
	router := group // domain/api/posts

	router.Get("/get/followings/feed", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetFollowingsFeed)
	router.Get("/get/subscriptions/feed", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetSubscriptionsFeed)
	router.Get("/user/archives", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetArchivedPosts)
	router.Get("/get/public/:profileId", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetPublicPosts)
	router.Get("/get/exclusive/:profileId", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetExclusivePosts)
}

func reactionsRouter(group fiber.Router) {
	router := group // domain/api/posts

	router.Post("/like/:postId", middleware.UserAuthHandler, postcontrollers.LikePost)
	router.Post("/dislike/:postId", middleware.UserAuthHandler, postcontrollers.DislikePost)
	router.Delete("/remove/like/:postId", middleware.UserAuthHandler, postcontrollers.RemoveLike)
	router.Delete("/remove/dislike/:postId", middleware.UserAuthHandler, postcontrollers.RemoveDislike)
}
