package middlewares

import (
	"ebupot-app/config"
	"ebupot-app/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CurrentUser(c *gin.Context) *models.User {
	session := sessions.Default(c)
	uid := session.Get("user_id")
	if uid == nil {
		return nil
	}

	var user models.User
	if err := config.DB.First(&user, uid).Error; err != nil {
		return nil
	}
	return &user
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil {
			c.Redirect(302, "/login?error=Session+berakhir,+silakan+login+kembali")
			c.Abort()
			return
		}
		c.Set("currentUser", user)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil {
			c.Redirect(302, "/login?error=Session+berakhir,+silakan+login+kembali")
			c.Abort()
			return
		}
		if user.Role != "admin" {
			c.Redirect(302, "/user/dashboard")
			c.Abort()
			return
		}
		c.Set("currentUser", user)
		c.Next()
	}
}

func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil {
			c.Redirect(302, "/login?error=Session+berakhir,+silakan+login+kembali")
			c.Abort()
			return
		}
		if user.Role != "user" {
			c.Redirect(302, "/admin/dashboard")
			c.Abort()
			return
		}
		c.Set("currentUser", user)
		c.Next()
	}
}
