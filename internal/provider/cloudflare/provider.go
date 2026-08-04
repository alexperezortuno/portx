package cloudflare

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"sync"

	"github.com/alexperezortuno/portx/internal/provider"
)

type Provider struct {
	name        string
	status      provider.Status
	mu          sync.RWMutex
	cancel      context.CancelFunc
	url         string
	port        int
	daemonErr   error
	daemonErrCh chan error
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

	port, err := parsePort(cfg.LocalAddr)
	if err != nil {
		return fmt.Errorf("invalid local address %q: %w", cfg.LocalAddr, err)
	}

	childCtx, cancel := context.WithCancel(ctx)

	resultCh := make(chan struct {
		url string
		err error
	}, 1)

	daemonErrCh := make(chan error, 1)

	p.daemonErrCh = daemonErrCh

	go func() {
		tunURL, err := startQuickTunnel(childCtx, port, "warn", daemonErrCh)
		resultCh <- struct {
			url string
			err error
		}{url: tunURL, err: err}
	}()

	go func() {
		select {
		case err := <-daemonErrCh:
			p.mu.Lock()
			p.daemonErr = err
			p.mu.Unlock()
			slog.Error("cloudflare tunnel daemon error", "err", err)
		case <-childCtx.Done():
		}
	}()

	select {
	case <-ctx.Done():
		cancel()
		return ctx.Err()
	case res := <-resultCh:
		if res.err != nil {
			cancel()
			return fmt.Errorf("creating cloudflare tunnel: %w", res.err)
		}
		p.url = res.url
		p.port = port
		p.cancel = cancel
		p.status = provider.StatusRunning
		return nil
	}
}

func (p *Provider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != provider.StatusRunning {
		return nil
	}

	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}

	p.url = ""
	p.status = provider.StatusStopped
	return nil
}

func (p *Provider) Status(ctx context.Context) (provider.Status, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.daemonErr != nil {
		slog.Error("cloudflare tunnel daemon error", "err", p.daemonErr)
		return provider.StatusError, p.daemonErr
	}
	return p.status, nil
}

func (p *Provider) URL() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.url
}

var _ provider.Provider = (*Provider)(nil)

func parsePort(localAddr string) (int, error) {
	host, portStr, err := net.SplitHostPort(localAddr)
	if err != nil {
		return 0, fmt.Errorf("cannot parse host:port: %w", err)
	}

	if host != "localhost" && host != "127.0.0.1" && host != "::1" {
		return 0, fmt.Errorf("cloudflare quick tunnel requires localhost, got %q", host)
	}

	u, err := url.Parse("http://" + localAddr)
	if err != nil {
		return 0, fmt.Errorf("cannot parse address: %w", err)
	}
	if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
		return 0, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	port, err := strconvAtoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port %q: %w", portStr, err)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port %d out of range", port)
	}

	return port, nil
}

func strconvAtoi(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid digit %q", string(c))
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
