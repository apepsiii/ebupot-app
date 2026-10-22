package config

import (
	"log"
	"log/slog"

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
		slog.Error("Gagal koneksi ke database SQLite", "error", err)
		log.Fatal("Gagal koneksi ke database SQLite: ", err)
	}

	err = database.AutoMigrate(&models.User{}, &models.Ebupot{})
	if err != nil {
		slog.Error("Gagal migrasi database", "error", err)
		log.Fatal("Gagal migrasi database: ", err)
	}

	// Isi nomor_dokumen default untuk data lama jika masih kosong
	var legacyEbupots []models.Ebupot
	database.Where("nomor_dokumen IS NULL OR nomor_dokumen = ''").Find(&legacyEbupots)
	for _, eb := range legacyEbupots {
		database.Model(&models.Ebupot{}).Where("id = ?", eb.ID).Update("nomor_dokumen", "DOC001")
	}

	DB = database
	seedAdmin()
	slog.Info("Database Connected & Migrated!")
}

func seedAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Gagal hash password admin", "error", err)
		log.Fatal("Gagal hash password admin: ", err)
	}

	admin := models.User{
		Username: "admin",
		Password: string(hashed),
		Name:     "Administrator",
		Role:     "admin",
	}

	if err := DB.Create(&admin).Error; err != nil {
		slog.Error("Gagal membuat akun admin", "error", err)
		log.Fatal("Gagal membuat akun admin: ", err)
	}

	slog.Info("Akun admin default dibuat", "username", "admin", "note", "ganti password default segera")
}
