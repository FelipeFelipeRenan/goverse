package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

var (
	Log *slog.Logger
)

type contextKey string

var RequestIDKey = contextKey("requestID")

func Init(level, serviceName string) {
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	Log = slog.New(handler).With(
		"service", serviceName,
		"env", os.Getenv("ENV"),
	)
}

func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

func WithContext(ctx context.Context) *slog.Logger {
	requestID, _ := ctx.Value(RequestIDKey).(string)
	return Log.With("request_id", requestID)
}
