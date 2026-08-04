package frp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/alexperezortuno/portx/internal/provider"
)

type Provider struct {
	name   string
	cfg    *Config
	svc    *Service
	status provider.Status
	mu     sync.RWMutex
}

func New(name string) *Provider {
	return &Provider{
		name:   name,
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

	localPort, err := parsePort(cfg.LocalAddr)
	if err != nil {
		return fmt.Errorf("invalid local address %q: %w", cfg.LocalAddr, err)
	}

	frpCfg := p.cfg
	if frpCfg == nil {
		return fmt.Errorf("FRP config not set")
	}

	frpCfg.LocalPort = localPort

	svc, err := StartService(ctx, frpCfg)
	if err != nil {
		return fmt.Errorf("starting frp service: %w", err)
	}

	p.svc = svc
	p.status = provider.StatusRunning

	go func() {
		url, err := svc.WaitForReady(ctx)
		if err != nil {
			slog.Error("frp tunnel error", "err", err)
			p.mu.Lock()
			p.status = provider.StatusError
			p.mu.Unlock()
			return
		}

		if url != "" {
			p.mu.Lock()
			p.svc.SetURL(url)
			p.mu.Unlock()
			slog.Info("frp tunnel ready", "url", url)
		}
	}()

	return nil
}

func (p *Provider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != provider.StatusRunning {
		return nil
	}

	if p.svc != nil {
		if err := p.svc.Stop(ctx); err != nil {
			slog.Error("stopping frp service", "err", err)
		}
		p.svc = nil
	}

	p.status = provider.StatusStopped
	return nil
}

func (p *Provider) Status(ctx context.Context) (provider.Status, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.svc != nil {
		svcStatus := p.svc.Status()
		switch svcStatus {
		case StatusError:
			return provider.StatusError, nil
		case StatusRunning:
			return provider.StatusRunning, nil
		}
	}

	return p.status, nil
}

func (p *Provider) URL() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.svc != nil {
		return p.svc.URL()
	}
	return ""
}

func (p *Provider) SetConfig(cfg *Config) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cfg = cfg
}

func parsePort(localAddr string) (int, error) {
	host, portStr, err := net.SplitHostPort(localAddr)
	if err != nil {
		return 0, err
	}

	if host == "" {
		return 0, fmt.Errorf("host cannot be empty in address %q", localAddr)
	}

	port := 0
	for _, c := range portStr {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid port %q", portStr)
		}
		port = port*10 + int(c-'0')
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port %d out of range", port)
	}

	return port, nil
}

var _ provider.Provider = (*Provider)(nil)
