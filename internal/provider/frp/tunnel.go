package frp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	frpclient "github.com/fatedier/frp/client"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/samber/lo"
)

type Service struct {
	cancel context.CancelFunc
	errCh  chan error
	url    string
	mu     sync.RWMutex
	status Status
	wg     sync.WaitGroup
}

type Status string

const (
	StatusStarting Status = "starting"
	StatusRunning  Status = "running"
	StatusStopped  Status = "stopped"
	StatusError    Status = "error"
)

func StartService(ctx context.Context, cfg *Config) (*Service, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	ctx, cancel := context.WithCancel(ctx)

	svc := &Service{
		cancel: cancel,
		errCh:  make(chan error, 1),
		status: StatusStarting,
	}

	common := buildCommonConfig(cfg)
	proxyCfgs := buildProxyConfigs(cfg)

	svc.wg.Add(1)
	go func() {
		defer svc.wg.Done()
		if err := runFRPService(ctx, common, proxyCfgs, cfg, svc.errCh); err != nil {
			select {
			case svc.errCh <- err:
			case <-ctx.Done():
			}
		}
	}()

	return svc, nil
}

func runFRPService(ctx context.Context, common *v1.ClientCommonConfig, proxyCfgs []v1.ProxyConfigurer, cfg *Config, errCh chan<- error) error {
	slog.Info("starting frp client", "server", fmt.Sprintf("%s:%d", common.ServerAddr, common.ServerPort), "proxy", cfg.ProxyName())

	svr, err := frpclient.NewService(frpclient.ServiceOptions{
		Common:      common,
		ProxyCfgs:   proxyCfgs,
		VisitorCfgs: nil,
	})
	if err != nil {
		return fmt.Errorf("creating frp service: %w", err)
	}

	go func() {
		<-ctx.Done()
		slog.Debug("frp service context cancelled, stopping")
		svr.GracefulClose(5 * time.Second)
	}()

	if err := svr.Run(ctx); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("frp service run error: %w", err)
	}

	return nil
}

func buildCommonConfig(cfg *Config) *v1.ClientCommonConfig {
	common := &v1.ClientCommonConfig{}
	common.Complete()

	common.ServerAddr = cfg.ServerAddr
	common.ServerPort = cfg.ServerPort
	common.Auth.Method = "token"
	common.Auth.Token = cfg.Token
	common.User = cfg.User
	common.Transport.Protocol = "tcp"
	common.Transport.TCPMux = lo.ToPtr(true)
	common.Transport.DialServerTimeout = 10
	common.Transport.DialServerKeepAlive = 30
	common.Transport.TCPMuxKeepaliveInterval = 20
	common.Transport.HeartbeatInterval = 30
	common.Transport.HeartbeatTimeout = 90
	common.Log.Level = "warn"
	common.Log.To = "console"

	if cfg.TLSEnable {
		common.Transport.TLS.Enable = lo.ToPtr(true)
		common.Transport.TLS.DisableCustomTLSFirstByte = lo.ToPtr(true)
	}

	return common
}

func buildProxyConfigs(cfg *Config) []v1.ProxyConfigurer {
	base := v1.ProxyBaseConfig{
		Name: cfg.ProxyName(),
	}

	var proxy v1.ProxyConfigurer

	switch cfg.ProxyType {
	case ProxyTypeTCP:
		proxy = &v1.TCPProxyConfig{
			ProxyBaseConfig: base,
			RemotePort:      cfg.RemotePort,
		}
		proxy.GetBaseConfig().LocalIP = "127.0.0.1"
		proxy.GetBaseConfig().LocalPort = cfg.LocalPort
	case ProxyTypeHTTP:
		proxy = &v1.HTTPProxyConfig{
			ProxyBaseConfig: base,
			DomainConfig: v1.DomainConfig{
				SubDomain:     cfg.Subdomain,
				CustomDomains: []string{cfg.CustomDomain},
			},
		}
		proxy.GetBaseConfig().LocalIP = "127.0.0.1"
		proxy.GetBaseConfig().LocalPort = cfg.LocalPort
	case ProxyTypeHTTPS:
		proxy = &v1.HTTPSProxyConfig{
			ProxyBaseConfig: base,
			DomainConfig: v1.DomainConfig{
				SubDomain:     cfg.Subdomain,
				CustomDomains: []string{cfg.CustomDomain},
			},
		}
		proxy.GetBaseConfig().LocalIP = "127.0.0.1"
		proxy.GetBaseConfig().LocalPort = cfg.LocalPort
	}

	if proxy != nil {
		return []v1.ProxyConfigurer{proxy}
	}

	return nil
}

func (s *Service) WaitForReady(ctx context.Context) (string, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case err := <-s.errCh:
			s.mu.Lock()
			s.status = StatusError
			s.mu.Unlock()
			return "", err
		case <-ticker.C:
			s.mu.RLock()
			if s.status == StatusRunning {
				url := s.url
				s.mu.RUnlock()
				return url, nil
			}
			s.mu.RUnlock()
		}
	}
}

func (s *Service) SetURL(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.url = url
	s.status = StatusRunning
}

func (s *Service) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = StatusError
}

func (s *Service) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Service) URL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.url
}

func (s *Service) Stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	s.mu.Lock()
	s.status = StatusStopped
	s.url = ""
	s.mu.Unlock()
	return nil
}
