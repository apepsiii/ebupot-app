package config

import (
	"log"

	"ebupot-app/models"

	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dbPath := "data/ebupot.db"
	if Cfg != nil && Cfg.Database.Path != "" {
		dbPath = Cfg.Database.Path
	}

	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke database SQLite: ", err)
	}

	err = database.AutoMigrate(&models.User{}, &models.Ebupot{})
	if err != nil {
		log.Fatal("Gagal migrasi database: ", err)
	}

	DB = database
	seedAdmin()
	log.Println("Database Connected & Migrated!")
}

func seedAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Gagal hash password admin: ", err)
	}

	admin := models.User{
		Username: "admin",
		Password: string(hashed),
		Name:     "Administrator",
		Role:     "admin",
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Fatal("Gagal membuat akun admin: ", err)
	}

	log.Println("Akun admin default dibuat (username: admin, password: admin123)")
}
