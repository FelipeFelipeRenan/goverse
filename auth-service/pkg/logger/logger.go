package logger

import (
	"log/slog"
	"os"
)

var (
	Info  *slog.Logger
	Error *slog.Logger
)

func Init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)

	Info = logger.With("component", "INFO")
	Error = logger.With("component", "ERROR")
}
