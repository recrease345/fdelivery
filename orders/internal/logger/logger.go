package logger

// временно пока go.work не сделан

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

func New() (*slog.Logger, error) {
	logLevel, err := getLogLevel()
	if err != nil {
		return nil, fmt.Errorf("2failed to initialize logger: %w", err)
	}

	options := slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, &options)
	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger, nil
}

func getLogLevel() (slog.Level, error) {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		return slog.LevelInfo, nil
	}

	levels := make(map[string]slog.Level)
	levels["debug"] = slog.LevelDebug
	levels["info"] = slog.LevelInfo
	levels["warn"] = slog.LevelWarn
	levels["error"] = slog.LevelError

	logLevelCorrect := false
	for k := range levels {
		if logLevel == k {
			logLevelCorrect = true
		}
	}

	if logLevelCorrect {
		return levels[logLevel], nil
	}

	return 0, fmt.Errorf("invalid LOG_LEVEL value: %w", errors.New(logLevel))
}
