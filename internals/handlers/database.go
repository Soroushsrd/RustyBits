package handlers

import (
	"RustyBits/internals/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// API routes
func (h *Handler) GetPostsJson(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 10
	offset := (page - 1) * limit

	var posts []models.Post
	result := h.DB.Where("published = ?", true).
		Preload("Tags").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts)

	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "failed to load posts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"page":  page,
	})
}

func (h *Handler) GetPostJson(c *gin.Context) {
	id := c.Param("id")
	var post models.Post
	result := h.DB.Preload("Tags").First(&post, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post Not Found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load the post"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// Admin routes
func (h *Handler) AdminDashboard(c *gin.Context) {
	var stats struct {
		TotalPosts     int64
		PublishedPosts int64
		DraftPosts     int64
		TotalTags      int64
	}

	h.DB.Model(&models.Post{}).Count(&stats.TotalPosts)

	h.DB.Model(&models.Post{}).Where("publushed = ?", true).Count(&stats.PublishedPosts)
	h.DB.Model(&models.Post{}).Where("published = ?", false).Count(&stats.DraftPosts)
	h.DB.Model(&models.Tag{}).Count(&stats.TotalTags)

	var recentPosts []models.Post

	h.DB.Preload("Tags").Order("created_at DESC").Limit(5).Find(&recentPosts)

	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"stats":       stats,
		"recentPosts": recentPosts,
		"title":       "Dashboard",
	})
}

func (h *Handler) AdminPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 20
	offset := (page - 1) * limit

	var posts []models.Post
	var total int64

	h.DB.Model(&models.Post{}).Count(&total)
	result := h.DB.Preload("Tags").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts)

	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load posts",
		})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.HTML(http.StatusOK, "admin/posts.html", gin.H{
		"posts":       posts,
		"currentPage": page,
		"totalPages":  totalPages,
		"hasNext":     page < totalPages,
		"hasPrev":     page > 1,
		"title":       "Manage Posts",
	})
}

func (h *Handler) NewPostForm(c *gin.Context) {
	var tags []models.Tag
	h.DB.Find(&tags)

	c.HTML(http.StatusOK, "admin/post-form.html", gin.H{
		"post":   models.Post{},
		"tags":   tags,
		"title":  "New Post",
		"action": "/admin/posts",
		"method": "Post",
	})
}

func (h *Handler) CreatePost(c *gin.Context) {

	var post models.Post

	if err := c.ShouldBind(&post); err != nil {
		var tags []models.Tag
		h.DB.Find(&tags)

		c.HTML(http.StatusBadRequest, "admin/post-form.html", gin.H{
			"post":  post,
			"tags":  tags,
			"error": err.Error(),
		})
		return
	}

	post.Slug = generateSlug(post.Title)

	tagNames := c.PostFormArray("tags")
	var tags []models.Tag
	for _, tagName := range tagNames {
		if tagName != "" {
			var tag models.Tag
			result := h.DB.Where("name = ?", tagName).First(&tag)
			if result.Error == gorm.ErrRecordNotFound {
				tag = models.Tag{Name: tagName}
				h.DB.Create(&tag)
			}
			tags = append(tags, tag)
		}
	}
	post.Tags = tags

	if err := h.DB.Create(&post).Error; err != nil {
		var allTags []models.Tag
		h.DB.Find(&allTags)
		c.HTML(http.StatusInternalServerError, "admin/post-form.html", gin.H{
			"post":  post,
			"tags":  allTags,
			"error": "Failed to create post",
		})
		return
	}

	// For HTMX requests, return the new post row
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Trigger", "postCreated")
		c.HTML(http.StatusOK, "admin/post-row.html", gin.H{"post": post})
		return
	}

	c.Redirect(http.StatusFound, "/admin/posts")
}

func (h *Handler) EditPostForm(c *gin.Context) {
	id := c.Param("id")
	var post models.Post
	result := h.DB.Preload("Tags").First(&post, id)
	if result.Error != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{
			"message": "Post not found",
		})
		return
	}

	var tags []models.Tag
	h.DB.Find(&tags)
	c.HTML(http.StatusOK, "admin/post-form.html", gin.H{
		"post":   post,
		"tags":   tags,
		"title":  "Edit Post",
		"action": fmt.Sprintf("/admin/posts/%d", post.ID),
		"method": "PATCH",
	})
}

func (h *Handler) UpodatePost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	if err := h.DB.Preload("Tags").First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Post Not Found",
		})
		return
	}

	if err := c.ShouldBind(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.Slug = generateSlug(post.Title)

	tagNames := c.PostFormArray("tags")
	var tags []models.Tag
	for _, tagName := range tagNames {
		if tagName != "" {
			var tag models.Tag
			result := h.DB.Where("name = ?", tagName).First(&tag)
			if result.Error == gorm.ErrRecordNotFound {
				tag = models.Tag{Name: tagName}
				h.DB.Create(&tag)

			}
			tags = append(tags, tag)
		}

	}

	h.DB.Model(&post).Association("Tags").Replace(tags)

	if err := h.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// For HTMX requests, return updated post
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Trigger", "postUpdated")
		c.HTML(http.StatusOK, "admin/post-row.html", gin.H{"post": post})
		return
	}

	c.Redirect(http.StatusFound, "/admin/posts")
}

func (h *Handler) DeletePost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	if err := h.DB.First(&post, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// Delete associations first
	h.DB.Model(&post).Association("Tags").Clear()

	if err := h.DB.Delete(&post).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// For HTMX requests, return empty response
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Trigger", "postDeleted")
		c.Status(http.StatusOK)
		return
	}

	c.Redirect(http.StatusFound, "/admin/posts")
}

func (h *Handler) TogglePublished(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	if err := h.DB.First(&post, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	post.Published = !post.Published
	h.DB.Save(&post)

	// Return updated status for HTMX
	c.HTML(http.StatusOK, "admin/post-status.html", gin.H{"post": post})
}

// Auth Routes

func (h *Handler) LoginForm(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login",
	})
}

func (h *Handler) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	var user models.User
	result := h.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Invalid credentials",
			"email": email,
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Invalid credentials",
			"email": email,
		})
		return
	}

	// Set session/cookie (you'll need to implement session management)
	setUserSession(c, user.ID)

	c.Redirect(http.StatusFound, "/admin")
}

func (h *Handler) Logout(c *gin.Context) {
	clearUserSession(c)
	c.Redirect(http.StatusFound, "/")
}

// Page handlers?
func (h *Handler) Home(c *gin.Context) {
	var posts []models.Post

	result := h.DB.Where("published = ?", true).
		Preload("Tags").
		Order("created_at DESC").
		Limit(5).
		Find(&posts)

	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "failed to load posts",
		})
		return
	}

	c.HTML(http.StatusOK, "home.html", gin.H{
		"posts": posts,
		"title": "Welcome To My Blog",
	})
}

func (h *Handler) GetPostsByTag(c *gin.Context) {
	tagName := c.Param("tag")
	page, _ := strconv.Atoi((c.DefaultQuery("page", "1")))
	limit := 10
	offset := (page - 1) * limit

	var posts []models.Post
	var total int64

	h.DB.Model(&models.Post{}).
		Joins("JOIN post_tags ON posts.id = post_tags.post_id").
		Joins("JOIN tags ON post_tags.tag_id = tags.id").
		Where("tags.name = ? AND posts.published = ?", tagName, true).
		Count(&total)

	result := h.DB.
		Joins("JOIN post_tags ON posts.id = post_tags.post_id").
		Joins("JOIN tags ON post_tags.tag_id = tags.id").
		Where("tags.name = ? AND posts.published = ?", tagName, true).
		Preload("Tags").
		Order("posts.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts)

	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "failed to load posts",
		})
		return
	}

	totalPAges := int((total + int64(limit) - 1) / int64(limit))

	c.HTML(http.StatusOK, "posts.html", gin.H{
		"posts":       posts,
		"currentPage": page,
		"totalPages":  totalPAges,
		"hasNext":     page < totalPAges,
		"hasPrev":     page > 1,
		"title":       fmt.Sprintf("Posts tagged: %s", tagName),
		"tag":         tagName,
	})
}

func (h *Handler) RSS(c *gin.Context) {
	var posts []models.Post

	result := h.DB.Where("published = ?", true).
		Order("created_at DESC").
		Limit(20).
		Find(&posts)

	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error generating RSS feed")
		return
	}

	c.Header("Content-Type", "application/rss+xml")
	c.HTML(http.StatusOK, "rss.xml", gin.H{
		"posts":     posts,
		"buildDate": time.Now().Format(time.RFC1123Z),
	})
}

func (h *Handler) GetPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 10
	offset := (page - 1) * limit

	var posts []models.Post
	var total int64

	h.DB.Where("published = ?", true).Count(&total)

	result := h.DB.Where("published = ?", true).
		Preload("Tags").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts)

	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load posts",
		})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.HTML(http.StatusOK, "posts.html", gin.H{
		"posts":       posts,
		"currentPage": page,
		"totalPages":  totalPages,
		"hasNext":     page < totalPages,
		"hasPrev":     page > 1,
		"title":       "All Posts",
	})
}
func (h *Handler) GetPost(c *gin.Context) {
	slug := c.Param("slug")
	var post models.Post

	result := h.DB.Where("slug = ? AND published = ?", slug, true).Preload("Tags").First(&post)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.HTML(http.StatusNotFound, "404.html", gin.H{
				"message": "Post Not Found",
			})
			return
		}
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load post",
		})
		return
	}

	c.HTML(http.StatusOK, "post.html", gin.H{
		"post":  post,
		"title": post.Title,
	})

}

// Helper functions

func generateSlug(title string) string {
	// Simple slug generation - you might want to use a proper library
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (basic implementation)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Session management helpers (basic implementation)
func setUserSession(c *gin.Context, userID uint) {
	// This is a basic implementation - use proper session management in production
	c.SetCookie("user_id", fmt.Sprintf("%d", userID), 3600*24*7, "/", "", false, true)
}

func clearUserSession(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "", false, true)
}

func getCurrentUserID(c *gin.Context) (uint, error) {
	userIDStr, err := c.Cookie("user_id")
	if err != nil {
		return 0, err
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(userID), nil
}
