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
	Scheme       string `yaml:"scheme"`        // http | https | kosong=auto-detect dari request
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

// EnvExists memeriksa apakah file .env sudah ada (first run detection).
func EnvExists() bool {
	_, err := os.Stat(".env")
	return err == nil
}

// Load membaca embedded config YAML, memuat .env, lalu override dari environment.
func Load(yamlData []byte) *Config {
	cfg := parseYAML(yamlData)

	// Muat .env jika ada (tidak fatal bila tidak ada)
	_ = godotenv.Load()

	applyEnvOverrides(cfg)

	Cfg = cfg

	if cfg.App.Env == "production" && strings.Contains(strings.ToLower(cfg.Session.Secret), "change-this") {
		log.Println("PERINGATAN: SESSION_SECRET masih default! Wajib diatur via .env untuk production.")
	}

	return cfg
}

func parseYAML(data []byte) *Config {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Gagal parse config YAML: %v", err)
	}
	return &cfg
}

func applyEnvOverrides(cfg *Config) {
	envStr("APP_ENV", &cfg.App.Env)

	envStr("SERVER_HOST", &cfg.Server.Host)
	envStr("SERVER_PORT", &cfg.Server.Port)
	envStr("SERVER_DOMAIN", &cfg.Server.Domain)
	envStr("SERVER_SCHEME", &cfg.Server.Scheme)
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

// Domain mengembalikan domain publik dari config.
func (c *Config) Domain() string {
	return c.Server.Domain
}

// Scheme mengembalikan scheme dari config, atau string kosong jika auto-detect.
func (c *Config) Scheme() string {
	return c.Server.Scheme
}
