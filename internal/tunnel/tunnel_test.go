package tunnel

import (
	"context"
	"testing"

	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTunnelProvider struct {
	name    string
	started bool
}

func (m *mockTunnelProvider) Name() string { return m.name }

func (m *mockTunnelProvider) Start(ctx context.Context, cfg provider.TunnelConfig) error {
	m.started = true
	return nil
}

func (m *mockTunnelProvider) Stop(ctx context.Context) error {
	m.started = false
	return nil
}

func (m *mockTunnelProvider) Status(ctx context.Context) (provider.Status, error) {
	if m.started {
		return provider.StatusRunning, nil
	}
	return provider.StatusStopped, nil
}

func TestNewManager(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)
}

func TestManager_Add(t *testing.T) {
	m := NewManager()
	p := &mockTunnelProvider{name: "test"}

	err := m.Add(context.Background(), "web", provider.TunnelConfig{
		LocalAddr:  "localhost:8080",
		RemoteAddr: "0.0.0.0:10000",
	}, p)
	require.NoError(t, err)

	tun, ok := m.Get(context.Background(), "web")
	assert.True(t, ok)
	assert.Equal(t, "web", tun.Name)
	assert.Equal(t, provider.StatusStopped, tun.Status)
}

func TestManager_AddDuplicate(t *testing.T) {
	m := NewManager()
	p := &mockTunnelProvider{name: "test"}

	err := m.Add(context.Background(), "web", provider.TunnelConfig{}, p)
	require.NoError(t, err)

	err = m.Add(context.Background(), "web", provider.TunnelConfig{}, p)
	assert.Error(t, err)
}

func TestManager_Remove(t *testing.T) {
	m := NewManager()
	p := &mockTunnelProvider{name: "test"}

	err := m.Add(context.Background(), "web", provider.TunnelConfig{}, p)
	require.NoError(t, err)

	err = m.Remove(context.Background(), "web")
	require.NoError(t, err)

	_, ok := m.Get(context.Background(), "web")
	assert.False(t, ok)
}

func TestManager_RemoveRunning(t *testing.T) {
	m := NewManager()
	p := &mockTunnelProvider{name: "test"}

	err := m.Add(context.Background(), "web", provider.TunnelConfig{}, p)
	require.NoError(t, err)

	_ = m.StartAll(context.Background())

	err = m.Remove(context.Background(), "web")
	require.NoError(t, err)
	assert.False(t, p.started)
}

func TestManager_RemoveNotFound(t *testing.T) {
	m := NewManager()

	err := m.Remove(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestManager_Get(t *testing.T) {
	m := NewManager()

	_, ok := m.Get(context.Background(), "nonexistent")
	assert.False(t, ok)
}

func TestManager_List(t *testing.T) {
	m := NewManager()

	list := m.List(context.Background())
	assert.Empty(t, list)

	p := &mockTunnelProvider{name: "test"}
	_ = m.Add(context.Background(), "web", provider.TunnelConfig{}, p)
	_ = m.Add(context.Background(), "api", provider.TunnelConfig{}, p)

	list = m.List(context.Background())
	assert.Len(t, list, 2)
}

func TestManager_StartAll(t *testing.T) {
	m := NewManager()
	p1 := &mockTunnelProvider{name: "test1"}
	p2 := &mockTunnelProvider{name: "test2"}

	_ = m.Add(context.Background(), "web", provider.TunnelConfig{}, p1)
	_ = m.Add(context.Background(), "api", provider.TunnelConfig{}, p2)

	err := m.StartAll(context.Background())
	require.NoError(t, err)

	assert.True(t, p1.started)
	assert.True(t, p2.started)

	tun, _ := m.Get(context.Background(), "web")
	assert.Equal(t, provider.StatusRunning, tun.Status)
}

func TestManager_StartAllAlreadyRunning(t *testing.T) {
	m := NewManager()
	p := &mockTunnelProvider{name: "test"}

	_ = m.Add(context.Background(), "web", provider.TunnelConfig{}, p)

	_ = m.StartAll(context.Background())
	err := m.StartAll(context.Background())
	assert.NoError(t, err)
}

func TestManager_StopAll(t *testing.T) {
	m := NewManager()
	p1 := &mockTunnelProvider{name: "test1"}
	p2 := &mockTunnelProvider{name: "test2"}

	_ = m.Add(context.Background(), "web", provider.TunnelConfig{}, p1)
	_ = m.Add(context.Background(), "api", provider.TunnelConfig{}, p2)

	_ = m.StartAll(context.Background())
	err := m.StopAll(context.Background())
	require.NoError(t, err)

	assert.False(t, p1.started)
	assert.False(t, p2.started)

	tun, _ := m.Get(context.Background(), "web")
	assert.Equal(t, provider.StatusStopped, tun.Status)
}
