package routes

import (
	"RustyBits/internals/handlers"
	"RustyBits/internals/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	r.GET("/", handlers.HomePageHandler)
	r.GET("/posts", handlers.GetPosts)
	r.GET("/posts/:slug", handlers.GetPost)
	r.GET("/posts/:tag", handlers.GetPostsByTag)

	api := r.Group("/api")
	{
		api.GET("/posts", handlers.GetPostsJson)
		api.GET("/posts/:id", handlers.GetPostJson)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired())
	{
		admin.GET("/", handlers.AdminDashboard)
		admin.GET("/posts", handlers.AdminPosts)
		admin.GET("/posts/new", handlers.NewPostForm)
		admin.POST("/posts", handlers.CreatePost)
		admin.GET("/posts/:id/edit", handlers.EditPostForm)
		admin.PATCH("/posts/:id", handlers.UpdatePost)
		admin.DELETE("/posts/:id", handlers.DeletePost)
		admin.PATCH("/posts/:id/toggle", handlers.TogglePublished)
	}

	r.GET("/login", handlers.LoginForm)
	r.POST("/login", handlers.Login)
	r.POST("/logout", handlers.Logout)

}
