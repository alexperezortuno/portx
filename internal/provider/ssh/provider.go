package ssh

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/alexperezortuno/portx/internal/sshutil"
)

type SSHProvider struct {
	name      string
	config    sshutil.Config
	client    *sshutil.Client
	status    provider.Status
	mu        sync.RWMutex
	forwarder *sshutil.Forward
}

func New(name string, cfg sshutil.Config) *SSHProvider {
	return &SSHProvider{
		name:   name,
		config: cfg,
		status: provider.StatusStopped,
	}
}

func (p *SSHProvider) Name() string { return p.name }

func (p *SSHProvider) Start(ctx context.Context, cfg provider.TunnelConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status == provider.StatusRunning {
		return nil
	}

	dialer := sshutil.NewDialer(&p.config)
	client, err := dialer.Dial()
	if err != nil {
		return err
	}

	remoteAddr := cfg.RemoteAddr
	if remoteAddr == "" {
		remoteAddr = p.config.RemoteAddr
	}

	fwd := &sshutil.Forward{
		Client:     client,
		RemoteAddr: remoteAddr,
		LocalAddr:  cfg.LocalAddr,
	}

	if err := fwd.Start(ctx); err != nil {
		client.Close()
		return fmt.Errorf("starting reverse forward: %w", err)
	}

	p.client = client
	p.forwarder = fwd
	p.status = provider.StatusRunning

	return nil
}

func (p *SSHProvider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != provider.StatusRunning {
		return nil
	}

	if p.forwarder != nil {
		p.forwarder.Stop()
		p.forwarder = nil
	}

	if p.client != nil {
		p.client.Close()
		p.client = nil
	}

	p.status = provider.StatusStopped
	return nil
}

func (p *SSHProvider) Status(ctx context.Context) (provider.Status, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status, nil
}

var _ provider.Provider = (*SSHProvider)(nil)
