package logger

import (
	"log/slog"
	"os"
)

func Init() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
}
