package main

import (
	"log"

	"ebupot-app/config"
	"ebupot-app/routes"
	"ebupot-app/setup"
)

func main() {
	var cfg *config.Config

	// Cek apakah .env sudah ada (first run detection)
	if !config.EnvExists() {
		// Jalankan setup wizard interaktif
		cfg = setup.RunWizard(configYAML)
	} else {
		// Muat config dari embedded YAML + .env
		cfg = config.Load(configYAML)
	}

	log.Printf("Aplikasi: %s | Env: %s | Listen: %s", cfg.App.Name, cfg.App.Env, cfg.Addr())

	// Inisialisasi Database
	config.ConnectDatabase()

	// Setup Router (templates & public dari embed.FS dalam binary)
	r := routes.SetupRouter(templateFS, publicFS)

	// Jalankan Server
	log.Printf("Server berjalan di http://localhost:%s (domain QR: %s)", cfg.Server.Port, cfg.Server.Domain)
	if err := r.Run(cfg.Addr()); err != nil {
		log.Fatal("Gagal menjalankan server: ", err)
	}
}
