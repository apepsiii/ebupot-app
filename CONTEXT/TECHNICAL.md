Struktur folder dan kode yang rapi sangat penting dalam pengembangan Go (Golang) agar proyek mudah dikelola (Maintainable), mudah di- *scale*, dan tidak membingungkan saat kode semakin banyak.

Untuk aplikasi berbasis Gin Framework + HTML Templates + SQLite, kita akan menggunakan pola **MVC (Model-View-Controller)** yang disesuaikan dengan ekosistem Go.

Berikut adalah rekomendasi struktur folder dan kode awal untuk proyek e-Bupot Anda:

### 1. Struktur Folder Proyek

Buat folder utama proyek Anda (misal: `ebupot-app`), lalu buat struktur di dalamnya seperti ini:

```text
ebupot-app/
│
├── config/                 # Konfigurasi aplikasi (Database, Environment)
│   └── database.go
│
├── controllers/            # Logika utama untuk menangani request dari browser
│   ├── auth_controller.go
│   ├── admin_controller.go
│   ├── user_controller.go
│   └── download_controller.go
│
├── middlewares/            # Fungsi penengah (proteksi rute/login session)
│   └── auth_middleware.go
│
├── models/                 # Struktur data (Struct) untuk GORM & SQLite
│   ├── user.go
│   └── ebupot.go
│
├── routes/                 # Kumpulan endpoint URL (Routing Gin)
│   └── routes.go
│
├── public/                 # File statis yang bisa diakses publik secara langsung
│   ├── css/
│   ├── js/
│   └── img/
│
├── templates/              # File HTML (Views)
│   ├── layout/
│   │   ├── header.html
│   │   └── footer.html
│   ├── auth/
│   │   └── login.html
│   ├── admin/
│   │   ├── dashboard.html
│   │   ├── users.html
│   │   └── ebupots.html
│   └── user/
│       └── dashboard.html
│
├── uploads/                # Folder TERPROTEKSI untuk menyimpan file PDF fisik
│   └── .gitkeep            # File kosong agar folder ikut ter-commit di Git
│
├── data/                   # Tempat menyimpan file database SQLite
│   └── .gitkeep            
│
├── main.go                 # File eksekusi utama aplikasi
├── go.mod                  # Manajemen dependency Go
└── go.sum                  # Checksum dependency

```

---

### 2. Contoh Kode Dasar (Boilerplate)

Agar Anda ada gambaran bagaimana file-file tersebut saling terhubung, berikut adalah contoh penulisan kode di beberapa file utama:

#### A. `models/user.go` & `models/ebupot.go`

Ini adalah representasi tabel database Anda menggunakan GORM.

**`models/user.go`**

```go
package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;not null"`
	Password  string    `gorm:"not null"`
	Name      string    `gorm:"not null"`
	Role      string    `gorm:"type:varchar(20);default:'user'"` // 'admin' atau 'user'
	CreatedAt time.Time
}

```

**`models/ebupot.go`**

```go
package models

import "time"

type Ebupot struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"` // Relasi ke User
	Bulan     int       `gorm:"not null"`
	Tahun     int       `gorm:"not null"`
	FilePath  string    `gorm:"not null"` // Contoh: uploads/ebupots/123.pdf
	UUIDLink  string    `gorm:"uniqueIndex;not null"` // ID unik untuk link
	CreatedAt time.Time
	
	User      User      `gorm:"foreignKey:UserID"` // Relasi untuk memudahkan Join
}

```

#### B. `config/database.go`

Berfungsi untuk membuka koneksi ke SQLite dan melakukan auto-migrasi tabel.

```go
package config

import (
	"log"
	"ebupot-app/models" // sesuaikan dengan nama module di go.mod Anda
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("data/ebupot.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke database SQLite!", err)
	}

	// Auto Migrate akan membuat tabel otomatis berdasarkan Struct di models
	err = database.AutoMigrate(&models.User{}, &models.Ebupot{})
	if err != nil {
		log.Fatal("Gagal migrasi database!", err)
	}

	DB = database
	log.Println("Database Connected & Migrated!")
}

```

#### C. `routes/routes.go`

Berfungsi memisahkan daftar URL (*routing*) agar `main.go` tidak terlalu panjang. Di sinilah *link* panjang untuk *download* didefinisikan.

```go
package routes

import (
	"ebupot-app/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 1. Setup tempat file HTML & Statis berada
	r.Static("/public", "./public")
	r.LoadHTMLGlob("templates/**/*")

	// 2. Rute Publik & Mesin Unduh Khusus (Sesuai Format DJP Coretax)
	r.GET("/", controllers.ShowLandingPage)
	r.GET("/login", controllers.ShowLogin)
	r.POST("/login", controllers.ProcessLogin)
	
	// Rute QR Code Target
	r.GET("/documentmanagementportal/api/DocumentExternalLink/:uuid", controllers.DirectDownloadHandler)

	// 3. Rute Dashboard (Nanti akan dibungkus dengan Middleware Session)
	adminGroup := r.Group("/admin")
	// adminGroup.Use(middlewares.RequireAdmin()) -> Akan kita buat nanti
	{
		adminGroup.GET("/dashboard", controllers.AdminDashboard)
		adminGroup.GET("/users", controllers.AdminUsers)
		// ... rute admin lainnya ...
	}

	userGroup := r.Group("/user")
	// userGroup.Use(middlewares.RequireUser()) -> Akan kita buat nanti
	{
		userGroup.GET("/dashboard", controllers.UserDashboard)
	}

	return r
}

```

#### D. `main.go`

Titik awal berjalannya aplikasi. Sangat singkat dan bersih.

```go
package main

import (
	"ebupot-app/config"
	"ebupot-app/routes"
)

func main() {
	// 1. Inisialisasi Database
	config.ConnectDatabase()

	// 2. Setup Router dari folder routes
	r := routes.SetupRouter()

	// 3. Jalankan Server di port 8080
	r.Run(":8080") 
}

```

### Langkah Apa Selanjutnya?

Dengan struktur di atas, setiap bagian aplikasi punya tugas yang jelas (*Separation of Concerns*). Jika ada *error* pada *database*, Anda cukup mengecek `config` atau `models`. Jika *error* tampilan, cek `templates`. Jika *error* URL, cek `routes` atau `controllers`.

Untuk memulai langkah nyatanya:

1. Buka Terminal/CMD.
2. Buat folder `ebupot-app` dan masuk ke dalamnya.
3. Jalankan `go mod init ebupot-app`.
4. Instal paket wajib dengan perintah:
`go get [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin) gorm.io/gorm gorm.io/driver/sqlite`
