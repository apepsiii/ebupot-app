package controllers

import (
	"fmt"
	"os"
	"strconv"

	"ebupot-app/config"
	"ebupot-app/models"

	"github.com/gin-gonic/gin"
)

func UserDashboard(c *gin.Context) {
	user := c.MustGet("currentUser").(*models.User)

	query := config.DB.Preload("User").Where("user_id = ?", user.ID)

	bulanStr := c.Query("bulan")
	tahunStr := c.Query("tahun")

	filterBulan := 0
	filterTahun := 0
	if bulanStr != "" {
		if b, err := strconv.Atoi(bulanStr); err == nil && b >= 1 && b <= 12 {
			query = query.Where("bulan = ?", b)
			filterBulan = b
		}
	}
	if tahunStr != "" {
		if t, err := strconv.Atoi(tahunStr); err == nil && t >= 1900 {
			query = query.Where("tahun = ?", t)
			filterTahun = t
		}
	}

	var ebupots []models.Ebupot
	query.Order("tahun DESC, bulan DESC").Find(&ebupots)

	// Daftar tahun yang tersedia untuk filter
	var years []int
	config.DB.Model(&models.Ebupot{}).Where("user_id = ?", user.ID).Distinct("tahun").Order("tahun DESC").Pluck("tahun", &years)

	c.HTML(200, "user_dashboard.html", gin.H{
		"title":       "Dashboard Saya",
		"active":      "dashboard",
		"ebupots":     ebupots,
		"filterBulan": filterBulan,
		"filterTahun": filterTahun,
		"years":       years,
		"currentUser": user,
	})
}

func UserEbupotDownload(c *gin.Context) {
	id := c.Param("id")
	user := c.MustGet("currentUser").(*models.User)

	var ebupot models.Ebupot
	if err := config.DB.First(&ebupot, id).Error; err != nil {
		c.HTML(404, "404.html", nil)
		return
	}

	// Isolasi data: pastikan dokumen milik user yang login
	if ebupot.UserID != user.ID {
		c.HTML(404, "404.html", nil)
		return
	}

	if _, err := os.Stat(ebupot.FilePath); os.IsNotExist(err) {
		c.HTML(404, "404.html", nil)
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", ebupot.FileName))
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(ebupot.FilePath)
}
