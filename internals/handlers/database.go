package handlers

import (
	"RustyBits/internals/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

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
