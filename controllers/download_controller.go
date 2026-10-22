package controllers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ebupot-app/config"
	"ebupot-app/models"

	"github.com/gin-gonic/gin"
)

// ExternalLinkRedirect mengarahkan link hasil scan QR ke halaman verifikasi dokumen.
// Sesuai format: /documentmanagementportal/api/DocumentExternalLink/:uuid
// -> redirect ke /documentmanagementportal/api/document-verify/:uuid
func ExternalLinkRedirect(c *gin.Context) {
	uuidParam := c.Param("uuid")

	var ebupot models.Ebupot
	if err := config.DB.Where("uuid_link = ?", uuidParam).First(&ebupot).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	c.Redirect(http.StatusFound, "/documentmanagementportal/api/document-verify/"+uuidParam)
}

// VerifyPage menampilkan formulir verifikasi nomor dokumen sebelum unduh.
func VerifyPage(c *gin.Context) {
	uuidParam := c.Param("uuid")

	var ebupot models.Ebupot
	if err := config.DB.Where("uuid_link = ?", uuidParam).First(&ebupot).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	errCode := c.Query("error")
	hasError := errCode != ""

	c.HTML(http.StatusOK, "verify.html", gin.H{
		"uuid":     uuidParam,
		"hasError": hasError,
		"error":    errCode,
	})
}

// VerifyAndDownload mencocokkan nomor dokumen yang diinput dengan data backend.
// Jika cocok: langsung trigger download file PDF (tidak menampilkan PDF di browser).
// Jika gagal: munculkan pesan error "Unduh Gagal - Nomor dokumen tidak sesuai."
func VerifyAndDownload(c *gin.Context) {
	uuidParam := c.Param("uuid")

	var ebupot models.Ebupot
	if err := config.DB.Where("uuid_link = ?", uuidParam).First(&ebupot).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	inputCode := strings.TrimSpace(c.PostForm("AccessCode"))
	if !matchNomorDokumen(&ebupot, inputCode) {
		c.Redirect(http.StatusFound, "/documentmanagementportal/api/document-verify/"+uuidParam+"?error=invalid_accesscode")
		return
	}

	if _, err := os.Stat(ebupot.FilePath); os.IsNotExist(err) {
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	fileName := ebupot.FileName
	if fileName == "" {
		fileName = filepath.Base(ebupot.FilePath)
	}

	// Trigger download file, tidak ada PDF yang ditampilkan di halaman browser
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	c.File(ebupot.FilePath)
}

// AltchaChallenge menghasilkan challenge proof-of-work ringan untuk altcha widget.
func AltchaChallenge(c *gin.Context) {
	saltBytes := make([]byte, 12)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)

	// Angka rahasia kecil agar verifikasi instan di sisi klien
	num := 5
	data := []byte(fmt.Sprintf("%s%d", salt, num))
	hash := sha256.Sum256(data)
	challenge := hex.EncodeToString(hash[:])

	c.JSON(http.StatusOK, gin.H{
		"algorithm": "SHA-256",
		"challenge": challenge,
		"maxnumber": 50,
		"salt":      salt,
		"signature": "",
	})
}

// matchNomorDokumen memeriksa kecocokan nomor dokumen secara fleksibel:
// 1. Field NomorDokumen spesifik
// 2. FileName lengkap (dengan atau tanpa .pdf)
// 3. Bagian komponen kode dokumen (seperti DOC001, dsb.)
func matchNomorDokumen(eb *models.Ebupot, input string) bool {
	clean := strings.TrimSpace(input)
	if clean == "" {
		return false
	}

	// 1. Cocokkan dengan NomorDokumen eksplisit
	if eb.NomorDokumen != "" && strings.EqualFold(clean, strings.TrimSpace(eb.NomorDokumen)) {
		return true
	}

	// 2. Cocokkan dengan nama file lengkap
	if strings.EqualFold(clean, strings.TrimSpace(eb.FileName)) {
		return true
	}

	// 3. Cocokkan tanpa ekstensi .pdf
	baseName := strings.TrimSuffix(eb.FileName, filepath.Ext(eb.FileName))
	if strings.EqualFold(clean, strings.TrimSpace(baseName)) {
		return true
	}

	// 4. Cocokkan segmen dalam nama file (misal: DOC001, BPA1_...)
	parts := strings.Split(baseName, "_")
	for _, p := range parts {
		if strings.EqualFold(clean, strings.TrimSpace(p)) {
			return true
		}
		subparts := strings.Split(p, "-")
		for _, sp := range subparts {
			if strings.EqualFold(clean, strings.TrimSpace(sp)) {
				return true
			}
		}
	}

	return false
}
