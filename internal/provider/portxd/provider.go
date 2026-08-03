package portxd

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexperezortuno/portx/internal/provider"
)

type Provider struct {
	name   string
	config ProviderConfig
	server *Server
	status provider.Status
	mu     sync.RWMutex
}

type ProviderConfig struct {
	RemoteAddr string
	LocalAddr  string
	ServerPort int
}

func New(name string, cfg ProviderConfig) *Provider {
	return &Provider{
		name:   name,
		config: cfg,
		status: provider.StatusStopped,
	}
}

func (p *Provider) Name() string { return p.name }

func (p *Provider) Start(ctx context.Context, cfg provider.TunnelConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == provider.StatusRunning {
		return nil
	}

	localAddr := cfg.LocalAddr
	if localAddr == "" {
		localAddr = p.config.LocalAddr
	}

	remoteAddr := cfg.RemoteAddr
	if remoteAddr == "" {
		remoteAddr = p.config.RemoteAddr
	}

	serverPort := p.config.ServerPort
	if serverPort == 0 {
		serverPort = 7222
	}

	srv := &Server{
		remoteAddr: remoteAddr,
		localAddr:  localAddr,
		serverPort: serverPort,
	}

	if err := srv.Start(ctx); err != nil {
		return fmt.Errorf("starting PortXD server: %w", err)
	}

	p.server = srv
	p.status = provider.StatusRunning

	return nil
}

func (p *Provider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != provider.StatusRunning {
		return nil
	}

	if p.server != nil {
		p.server.Stop()
		p.server = nil
	}

	p.status = provider.StatusStopped
	return nil
}

func (p *Provider) Status(ctx context.Context) (provider.Status, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status, nil
}

var _ provider.Provider = (*Provider)(nil)
