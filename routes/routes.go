package routes

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"ebupot-app/config"
	"ebupot-app/controllers"
	"ebupot-app/middlewares"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var monthsID = []string{
	"Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

func funcMap() template.FuncMap {
	return template.FuncMap{
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
	}
}

// loadTemplatesFromEmbed mem-parse semua template HTML dari embed.FS,
// menggunakan basename sebagai nama template (agar cocok dengan c.HTML).
func loadTemplatesFromEmbed(templateFS embed.FS) *template.Template {
	tmpl := template.New("").Funcs(funcMap())

	var walk func(dir string)
	walk = func(dir string) {
		es, err := fs.ReadDir(templateFS, dir)
		if err != nil {
			return
		}
		for _, e := range es {
			fullPath := dir + "/" + e.Name()
			if e.IsDir() {
				walk(fullPath)
				continue
			}
			if !strings.HasSuffix(e.Name(), ".html") {
				continue
			}
			data, err := templateFS.ReadFile(fullPath)
			if err != nil {
				panic("Gagal membaca template " + fullPath + ": " + err.Error())
			}
			name := filepath.Base(fullPath)
			_, err = tmpl.New(name).Parse(string(data))
			if err != nil {
				panic("Gagal parse template " + name + ": " + err.Error())
			}
		}
	}
	walk("templates")

	return tmpl
}

func SetupRouter(templateFS embed.FS, publicFS embed.FS) *gin.Engine {
	cfg := config.Cfg
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Session store (cookie-based) — secret dari config
	store := cookie.NewStore([]byte(cfg.Session.Secret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   cfg.Session.MaxAge,
		HttpOnly: true,
		Secure:   cfg.IsProduction(),
	})
	r.Use(sessions.Sessions("ebupot_session", store))

	// Templates dari embed.FS (di dalam binary, tidak perlu file di disk)
	tmpl := loadTemplatesFromEmbed(templateFS)
	r.SetHTMLTemplate(tmpl)

	// Static assets dari embed.FS
	publicSub, err := fs.Sub(publicFS, "public")
	if err != nil {
		panic("Gagal sub public FS: " + err.Error())
	}
	r.StaticFS("/public", http.FS(publicSub))

	// Rute Publik
	r.GET("/", controllers.RedirectToLogin)
	r.GET("/login", controllers.ShowLogin)
	r.POST("/login", controllers.ProcessLogin)
	r.GET("/logout", controllers.Logout)

	// Rute Mesin Unduh via QR Code (tanpa auth)
	r.GET("/documentmanagementportal/api/DocumentExternalLink/:uuid", controllers.DirectDownloadHandler)

	// Halaman 404 untuk rute tidak dikenal
	r.NoRoute(controllers.NotFound)

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
