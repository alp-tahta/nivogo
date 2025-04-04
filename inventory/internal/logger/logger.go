package logger

import (
	"log/slog"
	"os"
)

func Init() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	return logger
}
