package main

import (
	"log"

	"ebupot-app/config"
	"ebupot-app/routes"
)

func main() {
	// 1. Muat konfigurasi (config.yaml + .env)
	cfg := config.Load()
	log.Printf("Aplikasi: %s | Env: %s | Listen: %s", cfg.App.Name, cfg.App.Env, cfg.Addr())

	// 2. Inisialisasi Database
	config.ConnectDatabase()

	// 3. Setup Router
	r := routes.SetupRouter()

	// 4. Jalankan Server
	log.Printf("Server berjalan di http://localhost:%s (domain QR: %s)", cfg.Server.Port, cfg.Server.Domain)
	if err := r.Run(cfg.Addr()); err != nil {
		log.Fatal("Gagal menjalankan server: ", err)
	}
}
