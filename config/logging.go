package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

var Logger *slog.Logger

func InitLogging(logDir string) error {
	if logDir == "" {
		logDir = "logs"
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFile := filepath.Join(logDir, "app.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Tulis ke stdout DAN file sekaligus
	multi := io.MultiWriter(os.Stdout, f)

	Logger = slog.New(slog.NewJSONHandler(multi, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(Logger)
	return nil
}

func GetLogDir() string {
	return "logs"
}
