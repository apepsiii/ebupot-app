package routes

import (
	"ebupot-app/controllers"
	"ebupot-app/middlewares"
	"html/template"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var monthsID = []string{
	"Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

func setupFuncMap(r *gin.Engine) {
	r.SetFuncMap(template.FuncMap{
		"add1": func(i int) int { return i + 1 },
		"monthName": func(i int) string {
			if i >= 1 && i <= 12 {
				return monthsID[i-1]
			}
			return "-"
		},
		"monthsList": func() []int {
			return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		},
		"itoa": strconv.Itoa,
	})
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	setupFuncMap(r)

	// Session store (cookie-based)
	store := cookie.NewStore([]byte("ebupot-secret-key-change-in-production"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("ebupot_session", store))

	// Static files & templates
	r.Static("/public", "./public")
	r.LoadHTMLGlob("templates/**/*")

	// Rute Publik
	r.GET("/", controllers.ShowLandingPage)
	r.GET("/login", controllers.ShowLogin)
	r.POST("/login", controllers.ProcessLogin)
	r.GET("/logout", controllers.Logout)

	// Rute Mesin Unduh via QR Code (tanpa auth)
	r.GET("/documentmanagementportal/api/DocumentExternalLink/:uuid", controllers.DirectDownloadHandler)

	// Rute Dashboard Admin
	adminGroup := r.Group("/admin")
	adminGroup.Use(middlewares.RequireAdmin())
	{
		adminGroup.GET("/dashboard", controllers.AdminDashboard)

		adminGroup.GET("/users", controllers.AdminUsers)
		adminGroup.POST("/users", controllers.AdminUserCreate)
		adminGroup.POST("/users/update/:id", controllers.AdminUserUpdate)
		adminGroup.POST("/users/delete/:id", controllers.AdminUserDelete)

		adminGroup.GET("/ebupots", controllers.AdminEbupots)
		adminGroup.POST("/ebupots", controllers.AdminEbupotCreate)
		adminGroup.POST("/ebupots/update/:id", controllers.AdminEbupotUpdate)
		adminGroup.POST("/ebupots/delete/:id", controllers.AdminEbupotDelete)
		adminGroup.GET("/ebupots/download/:id", controllers.AdminEbupotDownload)
		adminGroup.GET("/ebupots/qr/:uuid", controllers.AdminEbupotQR)

		adminGroup.GET("/settings", controllers.AdminSettings)
		adminGroup.GET("/settings/logo/preview", controllers.AdminSettingsLogoPreview)
		adminGroup.POST("/settings/logo", controllers.AdminSettingsUploadLogo)
		adminGroup.POST("/settings/logo/delete", controllers.AdminSettingsDeleteLogo)
	}

	// Rute Dashboard User
	userGroup := r.Group("/user")
	userGroup.Use(middlewares.RequireUser())
	{
		userGroup.GET("/dashboard", controllers.UserDashboard)
		userGroup.GET("/ebupots/download/:id", controllers.UserEbupotDownload)
	}

	return r
}
