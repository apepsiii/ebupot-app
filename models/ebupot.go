package models

import "time"

type Ebupot struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Bulan     int       `gorm:"not null" json:"bulan"`
	Tahun     int       `gorm:"not null" json:"tahun"`
	FileName  string    `gorm:"not null" json:"file_name"`
	FilePath  string    `gorm:"not null" json:"file_path"`
	UUIDLink  string    `gorm:"uniqueIndex;not null" json:"uuid_link"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user"`
}
