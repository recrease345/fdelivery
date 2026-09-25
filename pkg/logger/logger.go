package logger

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	options := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewTextHandler(os.Stdout, &options)
	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}
