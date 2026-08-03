package logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger = *slog.Logger

type ctxKey struct{}

func New(ctx context.Context, level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.New(slog.NewJSONHandler(os.Stderr, opts))
	return handler
}

func NewDefault() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, nil))
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return NewDefault()
}

func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}
