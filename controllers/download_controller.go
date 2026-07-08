package controllers

import (
	"os"
	"path/filepath"

	"ebupot-app/config"
	"ebupot-app/models"

	"github.com/gin-gonic/gin"
)

func DirectDownloadHandler(c *gin.Context) {
	uuidParam := c.Param("uuid")

	var ebupot models.Ebupot
	if err := config.DB.Where("uuid_link = ?", uuidParam).First(&ebupot).Error; err != nil {
		c.HTML(404, "404.html", nil)
		return
	}

	if _, err := os.Stat(ebupot.FilePath); os.IsNotExist(err) {
		c.HTML(404, "404.html", nil)
		return
	}

	fileName := ebupot.FileName
	if fileName == "" {
		fileName = filepath.Base(ebupot.FilePath)
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(ebupot.FilePath)
}
