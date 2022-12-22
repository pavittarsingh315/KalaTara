package routes

import (
	"github.com/gofiber/fiber/v2"
	postcontrollers "nerajima.com/NeraJima/controllers/post_controllers"
	"nerajima.com/NeraJima/middleware"
)

func PostsRouter(group fiber.Router) {
	router := group.Group("/posts") // domain/api/posts

	crudRouter(router)
}

func crudRouter(group fiber.Router) {
	router := group

	router.Post("/create", middleware.UserAuthHandler, postcontrollers.CreatePost)
	router.Get("/get/:postId", middleware.UserAuthHandler, postcontrollers.GetPost)
	router.Put("/edit/:postId", middleware.UserAuthHandler, postcontrollers.EditPost)
	router.Delete("/delete/:postId", middleware.UserAuthHandler, postcontrollers.DeletePost)
}
