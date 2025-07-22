package logger

import (
	"context"
	"log/slog"
	"os"

	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
)

const (
	LevelFatal = slog.Level(10)
)

type Logger struct {
	Logger *slog.Logger
}

func NewLogger(config config.Config) *Logger {
	baseLogger := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.Level(config.LogLevel),
	})

	return &Logger{slog.New(baseLogger)}
}

func (m *Logger) Fatal(msg string, args ...interface{}) {
	m.Logger.Log(context.TODO(), LevelFatal, msg, args...)
	os.Exit(1)
}
