package logger

import (
	"log/slog"
	"os"

	"github.com/maiconpml/heylisten/internal/config"
)

var logFile *os.File

// Init configuress global logger to write to logging file.
func Init() error {
	var err error
	logPath, err := config.GetLogPath()
	if err != nil {
		return err
	}
	logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Sistema de log inicializado com sucesso")
	return nil
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
