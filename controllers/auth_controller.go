package controllers

import (
	"ebupot-app/config"
	"ebupot-app/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func ShowLogin(c *gin.Context) {
	errorMsg := c.Query("error")
	successMsg := c.Query("success")
	c.HTML(200, "login.html", gin.H{
		"title":   "e-Bupot Portal",
		"error":   errorMsg,
		"success": successMsg,
	})
}

// RedirectToLogin mengalihkan root ke halaman login.
func RedirectToLogin(c *gin.Context) {
	c.Redirect(302, "/login")
}

// NotFound menampilkan halaman 404 untuk rute yang tidak dikenal.
func NotFound(c *gin.Context) {
	c.HTML(404, "404.html", nil)
}

func ProcessLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.Redirect(302, "/login?error=Username+dan+Password+wajib+diisi")
		return
	}

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.Redirect(302, "/login?error=Username+atau+Password+salah")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.Redirect(302, "/login?error=Username+atau+Password+salah")
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("role", user.Role)
	if err := session.Save(); err != nil {
		c.Redirect(302, "/login?error=Gagal+menyimpan+sesi")
		return
	}

	if user.Role == "admin" {
		c.Redirect(302, "/admin/dashboard")
	} else {
		c.Redirect(302, "/user/dashboard")
	}
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(302, "/login?success=Anda+telah+berhasil+keluar")
}
