package logger

import (
	"log/slog"
	"os"
)

var (
	Info  *slog.Logger
	Error *slog.Logger
)

var (
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	reset  = "\033[0m"
)

func Init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)

	Info = logger.With("component", "INFO")
	Error = logger.With("component", "ERROR")
}
