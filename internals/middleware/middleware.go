package middleware

import (
	"RustyBits/internals/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getCurrentUserID(c)
		if err != nil {
			redirectToLogin(c)
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			clearUserSession(c)
			redirectToLogin(c)
			return
		}

		c.Set("user_id", userID)
		c.Set("user", user)
		c.Next()
	}
}

func OptionalAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getCurrentUserID(c)
		if err == nil {
			var user models.User
			if err := db.First(&user, userID).Error; err == nil {
				c.Set("user_id", userID)
				c.Set("user", user)
			}

		}
		c.Next()
	}
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

func clearUserSession(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "", false, true)
}

func RequireRole(db *gorm.DB, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			redirectToLogin(c)
			return
		}
		var user struct {
			ID   uint
			Role string
		}

		if err := db.Select("id,role").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied"})
			c.Abort()
			return
		}
		if user.Role != role && user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}
func isAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}
func redirectToLogin(c *gin.Context) {
	// for htmx reqs return a redirect header
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/login")
		c.Status(http.StatusUnauthorized)
		c.Abort()
		return
	}
	c.Redirect(http.StatusFound, "/login")
	c.Abort()
}
