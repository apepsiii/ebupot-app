package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"ebupot-app/config"
	"ebupot-app/routes"
	"ebupot-app/setup"
)

func main() {
	var cfg *config.Config
	firstRun := !config.EnvExists()

	if firstRun {
		cfg = setup.RunWizard(configYAML)
	} else {
		cfg = config.Load(configYAML)
	}

	logDir := cfg.Log.Dir
	if logDir == "" {
		logDir = "logs"
	}
	if err := config.InitLogging(logDir); err != nil {
		log.Fatalf("Gagal init logging: %v", err)
	}

	if firstRun {
		printPostSetupInfo(cfg)
	}

	slog.Info("Aplikasi started", "app", cfg.App.Name, "env", cfg.App.Env, "listen", cfg.Addr())

	config.ConnectDatabase()

	r := routes.SetupRouter(templateFS, publicFS)

	slog.Info("Server berjalan", "url", "http://"+cfg.Addr(), "domain_qr", cfg.Server.Domain)
	fmt.Println()
	fmt.Println(">>> Server berjalan di foreground. Tekan Ctrl+C untuk berhenti.")
	fmt.Println(">>> Untuk background: sudo systemctl start ebupot")
	fmt.Println()
	if err := r.Run(cfg.Addr()); err != nil {
		slog.Error("Gagal menjalankan server", "error", err)
		log.Fatal("Gagal menjalankan server: ", err)
	}
}

func printPostSetupInfo(cfg *config.Config) {
	wd, _ := os.Getwd()
	domain := cfg.Server.Domain
	scheme := "https"
	if cfg.Server.Scheme != "" {
		scheme = cfg.Server.Scheme
	}

	fmt.Println()
	fmt.Println("+======================================================+")
	fmt.Println("|              INFO APLIKASI                           |")
	fmt.Println("+======================================================+")
	fmt.Printf("|  Akses      : %s://%s\n", scheme, domain)
	fmt.Println("|  Login      : admin / admin123")
	fmt.Printf("|  .env       : %s\n", filepath.Join(wd, ".env"))
	fmt.Printf("|  Log file   : %s\n", filepath.Join(wd, "logs", "app.log"))
	fmt.Println("+======================================================+")
	fmt.Println()
	fmt.Println("Perintah berguna:")
	fmt.Println("  Mulai service : sudo systemctl start ebupot")
	fmt.Println("  Stop service  : sudo systemctl stop ebupot")
	fmt.Println("  Status service: sudo systemctl status ebupot")
	fmt.Println("  Lihat log     : sudo journalctl -u ebupot -f")
	fmt.Println("                 tail -f logs/app.log")
	fmt.Println()
	fmt.Println(">>> Memulai server...")
	fmt.Println()
}
