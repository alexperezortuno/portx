package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	name     string
	startFn  func(context.Context) error
	stopFn   func(context.Context) error
	statusFn func(context.Context) (Status, error)
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Start(ctx context.Context, cfg TunnelConfig) error {
	if m.startFn != nil {
		return m.startFn(ctx)
	}
	return nil
}

func (m *mockProvider) Stop(ctx context.Context) error {
	if m.stopFn != nil {
		return m.stopFn(ctx)
	}
	return nil
}

func (m *mockProvider) Status(ctx context.Context) (Status, error) {
	if m.statusFn != nil {
		return m.statusFn(ctx)
	}
	return StatusStopped, nil
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	p := &mockProvider{name: "test"}
	err := r.Register(p)
	require.NoError(t, err)

	got, ok := r.Get("test")
	assert.True(t, ok)
	assert.Equal(t, p, got)
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()

	p1 := &mockProvider{name: "test"}
	p2 := &mockProvider{name: "test"}

	err := r.Register(p1)
	require.NoError(t, err)

	err = r.Register(p2)
	assert.ErrorIs(t, err, ErrProviderAlreadyRegistered)
	assert.ErrorContains(t, err, "test")
}

func TestRegistry_RegisterNil(t *testing.T) {
	r := NewRegistry()

	err := r.Register(nil)
	assert.ErrorIs(t, err, ErrProviderNotFound)
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()

	p := &mockProvider{name: "test"}
	err := r.Register(p)
	require.NoError(t, err)

	err = r.Unregister("test")
	require.NoError(t, err)

	_, ok := r.Get("test")
	assert.False(t, ok)
}

func TestRegistry_UnregisterNotFound(t *testing.T) {
	r := NewRegistry()

	err := r.Unregister("nonexistent")
	assert.ErrorIs(t, err, ErrProviderNotFound)
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	_, ok := r.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	assert.Empty(t, r.List())

	err := r.Register(&mockProvider{name: "a"})
	require.NoError(t, err)
	err = r.Register(&mockProvider{name: "b"})
	require.NoError(t, err)

	list := r.List()
	assert.Len(t, list, 2)
	assert.Contains(t, list, "a")
	assert.Contains(t, list, "b")
}

func TestKnownProviders(t *testing.T) {
	providers := KnownProviders()
	assert.NotEmpty(t, providers)
	assert.Contains(t, providers, "ssh")
	assert.Contains(t, providers, "portxd")
	assert.Contains(t, providers, "cloudflare")
}

func TestIsKnown(t *testing.T) {
	assert.True(t, IsKnown("ssh"))
	assert.True(t, IsKnown("portxd"))
	assert.True(t, IsKnown("cloudflare"))
	assert.False(t, IsKnown(""))
}

func TestConfiguredProvider(t *testing.T) {
	startCalled := false
	stopCalled := false

	p := &mockProvider{
		name:     "test",
		startFn:  func(ctx context.Context) error { startCalled = true; return nil },
		stopFn:   func(ctx context.Context) error { stopCalled = true; return nil },
		statusFn: func(ctx context.Context) (Status, error) { return StatusRunning, nil },
	}

	cfg := ConfiguredProvider{
		Provider: p,
		Config: TunnelConfig{
			LocalAddr:  "localhost:8080",
			RemoteAddr: "0.0.0.0:10000",
		},
	}

	status, err := cfg.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, StatusRunning, status)

	startCalled = false
	stopCalled = false
	err = cfg.Start(context.Background())
	assert.NoError(t, err)
	assert.True(t, startCalled)

	err = cfg.Stop(context.Background())
	assert.NoError(t, err)
	assert.True(t, stopCalled)
}

func TestStatus_String(t *testing.T) {
	assert.Equal(t, "running", StatusRunning.String())
	assert.Equal(t, "stopped", StatusStopped.String())
	assert.Equal(t, "error", StatusError.String())
}

var (
	_ = errors.New
	_ = context.Background
)
