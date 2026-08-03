package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	l := New(ctx, slog.LevelInfo)

	assert.NotNil(t, l)
	assert.True(t, l.Handler().Enabled(context.Background(), slog.LevelInfo))
}

func TestNewDefault(t *testing.T) {
	l := NewDefault()
	assert.NotNil(t, l)
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()

	l := FromContext(ctx)
	assert.NotNil(t, l)

	ctx = context.WithValue(ctx, ctxKey{}, nil)
	l = FromContext(ctx)
	assert.NotNil(t, l)
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	l := NewDefault()

	ctx = WithContext(ctx, l)
	retrieved := FromContext(ctx)
	assert.Equal(t, l, retrieved)
}
