package routes

import (
	"RustyBits/internals/handlers"
	"RustyBits/internals/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	h := handlers.NewHandler(db)

	//  optional auth middleware to all routes to set user context if logged in
	r.Use(middleware.OptionalAuth(db))

	// Public routes
	r.GET("/", h.Home)
	r.GET("/posts/:slug", h.GetPost)
	r.GET("/posts", h.GetPosts)
	r.GET("/tags/:tag", h.GetPostsByTag)
	r.GET("/rss", h.RSS)

	//  routes for HTMX
	api := r.Group("/api")
	{
		api.GET("/posts", h.GetPostsJson)
		api.GET("/posts/:id", h.GetPostJson)
	}

	r.GET("/login", h.LoginForm)
	r.POST("/login", h.Login)
	r.POST("/logout", h.Logout)

	admin := r.Group("/admin")
	admin.Use(middleware.AuthRequired(db))
	{
		admin.GET("/", h.AdminDashboard)
		admin.GET("/posts", h.AdminPosts)
		admin.GET("/posts/new", h.NewPostForm)
		admin.POST("/posts", h.CreatePost)
		admin.GET("/posts/:id/edit", h.EditPostForm)
		admin.PATCH("/posts/:id", h.UpodatePost)
		admin.DELETE("/posts/:id", h.DeletePost)
		admin.PATCH("/posts/:id/toggle", h.TogglePublished)
	}
}

// func SetupAPIRoutes(r *gin.Engine, db *gorm.DB) {
// 	h := handlers.NewHandler(db)
//
// 	api := r.Group("/api/v1")
// 	{
// 		api.GET("/posts", h.GetPostsJson)
// 		api.GET("/posts/:id", h.GetPostJson)
//
// 		protected := api.Group("/")
// 		protected.Use(middleware.AuthRequired(db))
// 		{
// 			protected.POST("/posts", h.CreatePost)
// 			protected.PATCH("/posts/:id", h.UpodatePost)
// 			protected.DELETE("/posts/:id", h.DeletePost)
// 		}
// 	}
// }

func SetupRoutesWithCORS(r *gin.Engine, db *gorm.DB) {
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, HX-Request, HX-Target")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	SetupRoutes(r, db)
}
