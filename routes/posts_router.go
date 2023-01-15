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
	bookmarksRouter(router)
	commentsRouter(router)
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

	router.Get("/get/likes/:postId", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetLikesOfPost)
	router.Get("/get/dislikes/:postId", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetDislikesOfPost)

	router.Get("/liked/get", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetLikedPosts)
	router.Get("/disliked/get", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetDisikedPosts)
}

func bookmarksRouter(group fiber.Router) {
	router := group // domain/api/posts

	router.Post("/bookmark/:postId", middleware.UserAuthHandler, postcontrollers.BookmarkPost)
	router.Delete("/remove/bookmark/:postId", middleware.UserAuthHandler, postcontrollers.RemoveBookmark)
	router.Get("bookmarked/get", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetBookmarkedPosts)
}

func commentsRouter(group fiber.Router) {
	router := group.Group("/comments") // domain/api/posts/comments

	router.Get("/get/:postId", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetComments)
	router.Get("/:commentId/replies/get", middleware.UserAuthHandler, middleware.PaginationHandler, postcontrollers.GetReplies)

	router.Post("/create/:postId", middleware.UserAuthHandler, postcontrollers.CreateComment)
	router.Delete("/:commentId/remove", middleware.UserAuthHandler, postcontrollers.DeleteComment)
	router.Put("/:commentId/edit", middleware.UserAuthHandler, postcontrollers.EditComment)

	router.Post("/:commentId/like", middleware.UserAuthHandler, postcontrollers.LikeComment)
	router.Post("/:commentId/dislike", middleware.UserAuthHandler, postcontrollers.DislikeComment)
	router.Delete("/:commentId/like/remove", middleware.UserAuthHandler, postcontrollers.RemoveLikeFromComment)
	router.Delete("/:commentId/dislike/remove", middleware.UserAuthHandler, postcontrollers.RemoveDislikeFromComment)
}
