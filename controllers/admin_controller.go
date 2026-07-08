package controllers

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ebupot-app/config"
	"ebupot-app/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const maxUploadSize = 5 << 20 // 5 MB (default; override via config saat runtime)

func configuredMaxUploadSize() int64 {
	if config.Cfg != nil && config.Cfg.Upload.MaxSizeMB > 0 {
		return int64(config.Cfg.Upload.MaxSizeMB) << 20
	}
	return maxUploadSize
}

func AdminDashboard(c *gin.Context) {
	var userCount, ebupotCount int64
	config.DB.Model(&models.User{}).Count(&userCount)
	config.DB.Model(&models.Ebupot{}).Count(&ebupotCount)

	c.HTML(200, "admin_dashboard.html", gin.H{
		"title":        "Dashboard Admin",
		"active":       "dashboard",
		"userCount":    userCount,
		"ebupotCount":  ebupotCount,
		"currentUser":  c.MustGet("currentUser").(*models.User),
	})
}

func AdminUsers(c *gin.Context) {
	var users []models.User
	config.DB.Order("id DESC").Find(&users)

	c.HTML(200, "admin_users.html", gin.H{
		"title":       "Manajemen User",
		"active":      "users",
		"users":       users,
		"currentUser": c.MustGet("currentUser").(*models.User),
	})
}

func AdminUserCreate(c *gin.Context) {
	name := strings.TrimSpace(c.PostForm("name"))
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	role := c.PostForm("role")

	if name == "" || username == "" || password == "" {
		c.Redirect(302, "/admin/users?error=Semua+field+wajib+diisi")
		return
	}

	if role != "admin" && role != "user" {
		role = "user"
	}

	var existing models.User
	if config.DB.Where("username = ?", username).First(&existing).Error == nil {
		c.Redirect(302, "/admin/users?error=Username+sudah+digunakan")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.Redirect(302, "/admin/users?error=Gagal+mengenkripsi+password")
		return
	}

	user := models.User{
		Name:     name,
		Username: username,
		Password: string(hashed),
		Role:     role,
	}
	if err := config.DB.Create(&user).Error; err != nil {
		c.Redirect(302, "/admin/users?error=Gagal+menyimpan+user")
		return
	}

	c.Redirect(302, "/admin/users?success=User+berhasil+ditambahkan")
}

func AdminUserUpdate(c *gin.Context) {
	id := c.Param("id")
	name := strings.TrimSpace(c.PostForm("name"))
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	role := c.PostForm("role")

	if role != "admin" && role != "user" {
		role = "user"
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.Redirect(302, "/admin/users?error=User+tidak+ditemukan")
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)
	if user.ID == currentUser.ID && role != "admin" {
		c.Redirect(302, "/admin/users?error=Anda+tidak+dapat+mengubah+role+akun+sendiri")
		return
	}

	if name != "" {
		user.Name = name
	}
	if username != "" && username != user.Username {
		var existing models.User
		if config.DB.Where("username = ? AND id != ?", username, user.ID).First(&existing).Error == nil {
			c.Redirect(302, "/admin/users?error=Username+sudah+digunakan")
			return
		}
		user.Username = username
	}
	if password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.Redirect(302, "/admin/users?error=Gagal+mengenkripsi+password")
			return
		}
		user.Password = string(hashed)
	}
	user.Role = role

	if err := config.DB.Save(&user).Error; err != nil {
		c.Redirect(302, "/admin/users?error=Gagal+mengupdate+user")
		return
	}

	c.Redirect(302, "/admin/users?success=User+berhasil+diperbarui")
}

func AdminUserDelete(c *gin.Context) {
	id := c.Param("id")

	currentUser := c.MustGet("currentUser").(*models.User)
	if fmt.Sprintf("%d", currentUser.ID) == id {
		c.Redirect(302, "/admin/users?error=Anda+tidak+dapat+menghapus+akun+sendiri")
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.Redirect(302, "/admin/users?error=User+tidak+ditemukan")
		return
	}

	// Hapus file fisik ebupot milik user ini secara cascade
	var ebupots []models.Ebupot
	config.DB.Where("user_id = ?", user.ID).Find(&ebupots)
	for _, eb := range ebupots {
		if eb.FilePath != "" {
			os.Remove(eb.FilePath)
		}
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		c.Redirect(302, "/admin/users?error=Gagal+menghapus+user")
		return
	}

	c.Redirect(302, "/admin/users?success=User+berhasil+dihapus")
}

func AdminEbupots(c *gin.Context) {
	var ebupots []models.Ebupot
	config.DB.Preload("User").Order("id DESC").Find(&ebupots)

	var users []models.User
	config.DB.Where("role = ?", "user").Order("name ASC").Find(&users)

	c.HTML(200, "admin_ebupots.html", gin.H{
		"title":       "Manajemen e-Bupot",
		"active":      "ebupots",
		"ebupots":     ebupots,
		"users":       users,
		"currentUser": c.MustGet("currentUser").(*models.User),
	})
}

func AdminEbupotCreate(c *gin.Context) {
	maxSize := configuredMaxUploadSize()
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

	userIDStr := c.PostForm("user_id")
	bulanStr := c.PostForm("bulan")
	tahunStr := c.PostForm("tahun")

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.Redirect(302, "/admin/ebupots?error=User+wajib+dipilih")
		return
	}

	bulan, err := strconv.Atoi(bulanStr)
	if err != nil || bulan < 1 || bulan > 12 {
		c.Redirect(302, "/admin/ebupots?error=Bulan+tidak+valid")
		return
	}

	tahun, err := strconv.Atoi(tahunStr)
	if err != nil || tahun < 1900 {
		c.Redirect(302, "/admin/ebupots?error=Tahun+tidak+valid")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.Redirect(302, "/admin/ebupots?error=File+wajib+diunggah")
		return
	}

	// Validasi ekstensi PDF
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		c.Redirect(302, "/admin/ebupots?error=Hanya+file+berformat+PDF+yang+diizinkan")
		return
	}

	// Validasi tipe MIME
	if file.Header.Get("Content-Type") != "application/pdf" {
		buf := make([]byte, 4)
		f, _ := file.Open()
		f.Read(buf)
		f.Close()
		if !(buf[0] == 0x25 && buf[1] == 0x50 && buf[2] == 0x44 && buf[3] == 0x46) {
			c.Redirect(302, "/admin/ebupots?error=Hanya+file+berformat+PDF+yang+diizinkan")
			return
		}
	}

	// Validasi ukuran
	if file.Size > maxSize {
		c.Redirect(302, "/admin/ebupots?error=Ukuran+file+maksimal+"+strconv.Itoa(config.Cfg.Upload.MaxSizeMB)+"MB")
		return
	}

	newUUID := uuid.New().String()

	// Hitung nomor dokumen untuk user ini (DOC ke berapa)
	var docCount int64
	config.DB.Model(&models.Ebupot{}).Where("user_id = ?", userID).Count(&docCount)
	docNumber := int(docCount) + 1

	// Generate nama file sesuai format kustom
	customName := generateEbupotFileName(uint(userID), docNumber, bulan, tahun)
	storedName := newUUID + "_" + customName
	uploadDir := "uploads/ebupots"
	if config.Cfg != nil && config.Cfg.Upload.Dir != "" {
		uploadDir = config.Cfg.Upload.Dir
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.Redirect(302, "/admin/ebupots?error=Gagal+membuat+folder+upload")
		return
	}
	savePath := filepath.Join(uploadDir, storedName)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.Redirect(302, "/admin/ebupots?error=Gagal+menyimpan+file")
		return
	}

	ebupot := models.Ebupot{
		UserID:   uint(userID),
		Bulan:    bulan,
		Tahun:    tahun,
		FileName: customName,
		FilePath: savePath,
		UUIDLink: newUUID,
	}

	if err := config.DB.Create(&ebupot).Error; err != nil {
		os.Remove(savePath)
		c.Redirect(302, "/admin/ebupots?error=Gagal+menyimpan+data+ke+database")
		return
	}

	c.Redirect(302, "/admin/ebupots?success=Dokumen+berhasil+diunggah")
}

func AdminEbupotUpdate(c *gin.Context) {
	id := c.Param("id")

	var ebupot models.Ebupot
	if err := config.DB.First(&ebupot, id).Error; err != nil {
		c.Redirect(302, "/admin/ebupots?error=Dokumen+tidak+ditemukan")
		return
	}

	userIDStr := c.PostForm("user_id")
	bulanStr := c.PostForm("bulan")
	tahunStr := c.PostForm("tahun")

	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err == nil {
			ebupot.UserID = uint(userID)
		}
	}
	if bulanStr != "" {
		if b, err := strconv.Atoi(bulanStr); err == nil && b >= 1 && b <= 12 {
			ebupot.Bulan = b
		}
	}
	if tahunStr != "" {
		if t, err := strconv.Atoi(tahunStr); err == nil && t >= 1900 {
			ebupot.Tahun = t
		}
	}

	if err := config.DB.Save(&ebupot).Error; err != nil {
		c.Redirect(302, "/admin/ebupots?error=Gagal+mengupdate+dokumen")
		return
	}

	c.Redirect(302, "/admin/ebupots?success=Dokumen+berhasil+diperbarui")
}

func AdminEbupotDelete(c *gin.Context) {
	id := c.Param("id")

	var ebupot models.Ebupot
	if err := config.DB.First(&ebupot, id).Error; err != nil {
		c.Redirect(302, "/admin/ebupots?error=Dokumen+tidak+ditemukan")
		return
	}

	if ebupot.FilePath != "" {
		os.Remove(ebupot.FilePath)
	}

	if err := config.DB.Delete(&ebupot).Error; err != nil {
		c.Redirect(302, "/admin/ebupots?error=Gagal+menghapus+dokumen")
		return
	}

	c.Redirect(302, "/admin/ebupots?success=Dokumen+berhasil+dihapus")
}

func AdminEbupotQR(c *gin.Context) {
	uuidParam := c.Param("uuid")

	var ebupot models.Ebupot
	if err := config.DB.Where("uuid_link = ?", uuidParam).First(&ebupot).Error; err != nil {
		c.JSON(404, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	baseURL := getBaseURL(c)
	downloadURL := fmt.Sprintf("%s/documentmanagementportal/api/DocumentExternalLink/%s", baseURL, ebupot.UUIDLink)

	// Recovery level & ukuran dari config (high disarankan bila pakai logo)
	recoveryLevel := qrRecoveryLevel()
	qrSize := 512
	if config.Cfg != nil && config.Cfg.QR.Size > 0 {
		qrSize = config.Cfg.QR.Size
	}
	pngData, err := qrcode.Encode(downloadURL, recoveryLevel, qrSize)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal generate QR Code"})
		return
	}

	// Tempelkan logo di tengah jika tersedia
	pngData = overlayLogoOnQR(pngData)

	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"qr_%s.png\"", ebupot.UUIDLink))
	c.Data(200, "image/png", pngData)
}

// qrRecoveryLevel mengembalikan level koreksi error dari config.
func qrRecoveryLevel() qrcode.RecoveryLevel {
	lvl := "high"
	if config.Cfg != nil && config.Cfg.QR.RecoveryLevel != "" {
		lvl = strings.ToLower(config.Cfg.QR.RecoveryLevel)
	}
	switch lvl {
	case "low":
		return qrcode.Low
	case "medium":
		return qrcode.Medium
	case "highest":
		return qrcode.Highest
	default:
		return qrcode.High
	}
}

func getBaseURL(c *gin.Context) string {
	// Gunakan domain dari config (utama) agar URL QR stabil & konsisten
	if config.Cfg != nil && config.Cfg.Server.Domain != "" {
		return config.Cfg.BaseURL()
	}
	// Fallback: deteksi dari request
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if cf := c.GetHeader("X-Forwarded-Proto"); cf != "" {
		scheme = cf
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}

func AdminEbupotDownload(c *gin.Context) {
	id := c.Param("id")

	var ebupot models.Ebupot
	if err := config.DB.First(&ebupot, id).Error; err != nil {
		c.Redirect(302, "/admin/ebupots?error=Dokumen+tidak+ditemukan")
		return
	}

	if _, err := os.Stat(ebupot.FilePath); os.IsNotExist(err) {
		c.Redirect(302, "/admin/ebupots?error=File+fisik+tidak+ditemukan")
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", ebupot.FileName))
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(ebupot.FilePath)
}

// helper untuk menghindari import cycle pada user controller
func sessionUserID(c *gin.Context) uint {
	session := sessions.Default(c)
	uid := session.Get("user_id")
	if uid == nil {
		return 0
	}
	switch v := uid.(type) {
	case uint:
		return v
	case uint64:
		return uint(v)
	default:
		return 0
	}
}

var _ = gorm.ErrRecordNotFound

// AdminSettings menampilkan halaman pengaturan (logo QR).
func AdminSettings(c *gin.Context) {
	c.HTML(200, "admin_settings.html", gin.H{
		"title":       "Pengaturan",
		"active":      "settings",
		"hasLogo":     logoExists(),
		"currentUser": c.MustGet("currentUser").(*models.User),
	})
}

// AdminSettingsUploadLogo menangani upload logo untuk QR Code.
func AdminSettingsUploadLogo(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20) // 2MB

	file, err := c.FormFile("logo")
	if err != nil {
		c.Redirect(302, "/admin/settings?error=Logo+wajib+diunggah")
		return
	}

	// Validasi tipe gambar
	contentType := file.Header.Get("Content-Type")
	allowed := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/jpg":  true,
		"image/gif":  true,
	}
	if !allowed[contentType] {
		c.Redirect(302, "/admin/settings?error=Hanya+file+gambar+(PNG/JPG/GIF)+yang+diizinkan")
		return
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		c.Redirect(302, "/admin/settings?error=Gagal+menyiapkan+folder")
		return
	}

	// Simpan sementara lalu re-encode sebagai PNG untuk normalisasi
	tmpPath := filepath.Join("uploads", "_logo_tmp")
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		c.Redirect(302, "/admin/settings?error=Gagal+menyimpan+logo")
		return
	}

	// Decode gambar apa adanya, lalu encode ulang ke PNG
	f, err := os.Open(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		c.Redirect(302, "/admin/settings?error=Gagal+membaca+logo")
		return
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		os.Remove(tmpPath)
		c.Redirect(302, "/admin/settings?error=Format+gambar+tidak+valid")
		return
	}

	out, err := os.Create(logoPath())
	if err != nil {
		os.Remove(tmpPath)
		c.Redirect(302, "/admin/settings?error=Gagal+menyimpan+logo")
		return
	}
	defer out.Close()

	if err := png.Encode(out, img); err != nil {
		os.Remove(tmpPath)
		os.Remove(logoPath())
		c.Redirect(302, "/admin/settings?error=Gagal+memproses+logo")
		return
	}

	os.Remove(tmpPath)
	c.Redirect(302, "/admin/settings?success=Logo+berhasil+diunggah")
}

// AdminSettingsLogoPreview menampilkan logo yang tersimpan (untuk pratinjau).
func AdminSettingsLogoPreview(c *gin.Context) {
	if !logoExists() {
		c.Status(404)
		return
	}
	c.Header("Content-Type", "image/png")
	c.File(logoPath())
}

// AdminSettingsDeleteLogo menghapus logo yang tersimpan.
func AdminSettingsDeleteLogo(c *gin.Context) {
	if logoExists() {
		os.Remove(logoPath())
	}
	c.Redirect(302, "/admin/settings?success=Logo+berhasil+dihapus")
}
