package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config merepresentasikan seluruh pengaturan aplikasi.
type Config struct {
	App     AppConfig     `yaml:"app"`
	Server  ServerConfig  `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Upload  UploadConfig  `yaml:"upload"`
	QR      QRConfig      `yaml:"qr"`
	Session SessionConfig `yaml:"session"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Domain       string `yaml:"domain"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type UploadConfig struct {
	MaxSizeMB int    `yaml:"max_size_mb"`
	Dir       string `yaml:"dir"`
	LogoPath  string `yaml:"logo_path"`
}

type QRConfig struct {
	RecoveryLevel string `yaml:"recovery_level"`
	Size          int    `yaml:"size"`
}

type SessionConfig struct {
	Secret string `yaml:"secret"`
	MaxAge int    `yaml:"max_age"`
}

// Cfg adalah instance konfigurasi global.
var Cfg *Config

// Load membaca config.yaml, memuat .env, lalu meng-override nilai dari environment variables.
func Load() *Config {
	cfg := loadYAML("config.yaml")

	// Muat .env jika ada (tidak fatal bila tidak ada)
	_ = godotenv.Load()

	applyEnvOverrides(cfg)

	Cfg = cfg

	if cfg.App.Env == "production" && strings.Contains(strings.ToLower(cfg.Session.Secret), "change-this") {
		log.Println("PERINGATAN: SESSION_SECRET masih default! Wajib diatur via .env untuk production.")
	}

	return cfg
}

func loadYAML(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Gagal membaca %s: %v (pastikan file config.yaml ada)", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Gagal parse %s: %v", path, err)
	}
	return &cfg
}

func applyEnvOverrides(cfg *Config) {
	envStr("APP_ENV", &cfg.App.Env)

	envStr("SERVER_HOST", &cfg.Server.Host)
	envStr("SERVER_PORT", &cfg.Server.Port)
	envStr("SERVER_DOMAIN", &cfg.Server.Domain)
	envInt("SERVER_READ_TIMEOUT", &cfg.Server.ReadTimeout)
	envInt("SERVER_WRITE_TIMEOUT", &cfg.Server.WriteTimeout)

	envStr("DB_PATH", &cfg.Database.Path)

	envInt("UPLOAD_MAX_SIZE_MB", &cfg.Upload.MaxSizeMB)
	envStr("UPLOAD_DIR", &cfg.Upload.Dir)
	envStr("UPLOAD_LOGO_PATH", &cfg.Upload.LogoPath)

	envStr("QR_RECOVERY_LEVEL", &cfg.QR.RecoveryLevel)
	envInt("QR_SIZE", &cfg.QR.Size)

	envStr("SESSION_SECRET", &cfg.Session.Secret)
	envInt("SESSION_MAX_AGE", &cfg.Session.MaxAge)
}

// envStr meng-override pointer string bila env var ada.
func envStr(key string, target *string) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		*target = v
	}
}

// envInt meng-override pointer int bila env var ada & valid.
func envInt(key string, target *int) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*target = n
		}
	}
}

// Addr mengembalikan alamat listen (host:port).
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// IsProduction mengecek apakah env = production.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "production")
}

// BaseURL mengembalikan URL dasar untuk QR Code (http/https + domain).
func (c *Config) BaseURL() string {
	scheme := "https"
	if !c.IsProduction() {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Server.Domain)
}
