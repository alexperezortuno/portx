package tunnel

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/alexperezortuno/portx/internal/provider"
)

type TunnelManager interface {
	Add(ctx context.Context, name string, cfg provider.TunnelConfig, p provider.Provider) error
	Remove(ctx context.Context, name string) error
	Get(ctx context.Context, name string) (*Tunnel, bool)
	List(ctx context.Context) []*Tunnel
	StartAll(ctx context.Context) error
	StopAll(ctx context.Context) error
}

type Tunnel struct {
	Name     string
	Config   provider.TunnelConfig
	Provider provider.Provider
	Status   provider.Status
}

type manager struct {
	mu      sync.RWMutex
	tunnels map[string]*Tunnel
}

func NewManager() TunnelManager {
	return &manager{
		tunnels: make(map[string]*Tunnel),
	}
}

func (m *manager) Add(ctx context.Context, name string, cfg provider.TunnelConfig, p provider.Provider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tunnels[name]; exists {
		return fmt.Errorf("tunnel %q already exists", name)
	}

	m.tunnels[name] = &Tunnel{
		Name:     name,
		Config:   cfg,
		Provider: p,
		Status:   provider.StatusStopped,
	}

	return nil
}

func (m *manager) Remove(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, exists := m.tunnels[name]
	if !exists {
		return fmt.Errorf("tunnel %q not found", name)
	}

	if t.Status == provider.StatusRunning {
		if err := t.Provider.Stop(ctx); err != nil {
			return fmt.Errorf("stopping tunnel %q: %w", name, err)
		}
	}

	delete(m.tunnels, name)
	return nil
}

func (m *manager) Get(ctx context.Context, name string) (*Tunnel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, ok := m.tunnels[name]
	return t, ok
}

func (m *manager) List(ctx context.Context) []*Tunnel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tunnels := make([]*Tunnel, 0, len(m.tunnels))
	for _, t := range m.tunnels {
		tunnels = append(tunnels, t)
	}
	return tunnels
}

func (m *manager) StartAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, t := range m.tunnels {
		if t.Status == provider.StatusRunning {
			continue
		}
		if err := t.Provider.Start(ctx, t.Config); err != nil {
			errs = append(errs, fmt.Errorf("starting tunnel %q: %w", t.Name, err))
			t.Status = provider.StatusError
			continue
		}
		t.Status = provider.StatusRunning
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (m *manager) StopAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, t := range m.tunnels {
		if t.Status != provider.StatusRunning {
			continue
		}
		if err := t.Provider.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stopping tunnel %q: %w", t.Name, err))
			continue
		}
		t.Status = provider.StatusStopped
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

var _ TunnelManager = (*manager)(nil)
