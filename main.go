package main

import (
	"RustyBits/internals/models"
	"RustyBits/internals/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

	db, err := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	err = db.AutoMigrate(&models.Post{}, &models.Tag{}, &models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database", err)
	}

	createDefaultUser(db)

	// Initialize Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/**/*")

	// Serve static files
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")

	// Setup routes
	routes.SetupRoutes(r, db)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(r.Run(":" + port))
}

func createDefaultUser(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)

	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash password", err)
		}
		user := models.User{
			Email:    "admin@example.com",
			Password: string(hashedPassword),
		}

		if err := db.Create(&user).Error; err != nil {
			log.Fatal("Failed to create default user:", err)
		}

		log.Println("Created default admin user:")
		log.Println("Email: admin@example.com")
		log.Println("Password: admin123")
		log.Println("Please change these credentials after first login!")
	}
}

// Database seeding function (optional)
func seedDatabase(db *gorm.DB) {
	// Check if we already have posts
	var postCount int64
	db.Model(&models.Post{}).Count(&postCount)

	if postCount > 0 {
		return // Already seeded
	}

	// Create some sample tags
	tags := []models.Tag{
		{Name: "go"},
		{Name: "web-development"},
		{Name: "tutorial"},
		{Name: "programming"},
	}

	for _, tag := range tags {
		db.Create(&tag)
	}

	// Create sample posts
	posts := []models.Post{
		{
			Title:     "Welcome to My Blog",
			Slug:      "welcome-to-my-blog",
			Content:   "This is my first blog post. Welcome to my personal blog where I'll share my thoughts on programming, technology, and life.",
			Excerpt:   "Welcome to my personal blog where I'll share thoughts on programming and technology.",
			Published: true,
			Tags:      []models.Tag{tags[0], tags[1]}, // go, web-development
		},
		{
			Title:     "Getting Started with Go and HTMX",
			Slug:      "getting-started-go-htmx",
			Content:   "In this post, I'll show you how to build modern web applications using Go and HTMX. HTMX allows you to access AJAX, CSS Transitions, WebSockets and Server Sent Events directly in HTML.",
			Excerpt:   "Learn how to build modern web applications using Go and HTMX.",
			Published: true,
			Tags:      []models.Tag{tags[0], tags[2], tags[3]}, // go, tutorial, programming
		},
		{
			Title:     "Draft Post Example",
			Slug:      "draft-post-example",
			Content:   "This is a draft post that's not published yet.",
			Excerpt:   "An example of a draft post.",
			Published: false,
			Tags:      []models.Tag{tags[3]}, // programming
		},
	}

	for _, post := range posts {
		db.Create(&post)
	}

	log.Println("Database seeded with sample data")
}
