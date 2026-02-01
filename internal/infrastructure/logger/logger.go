package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

func New(level string) *Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return &Logger{
			Logger: l.With("request_id", requestID),
		}
	}
	return l
}

func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{
		Logger: l.With(key, value),
	}
}

func (l *Logger) WithFields(fields map[string]any) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.With(args...),
	}
}
