package runtime

import (
	"context"
	"log/slog"

	"github.com/alexperezortuno/portx/internal/config"
	"github.com/alexperezortuno/portx/internal/logger"
	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/alexperezortuno/portx/internal/tunnel"
	"github.com/alexperezortuno/portx/internal/version"
)

type Runtime struct {
	BuildInfo *version.BuildInfo
	Logger    *slog.Logger
	Config    *config.Config
	Registry  provider.ProviderRegistry
	Manager   tunnel.TunnelManager
}

type Option func(*Runtime)

func WithLogger(l *slog.Logger) Option {
	return func(r *Runtime) { r.Logger = l }
}

func WithConfig(cfg *config.Config) Option {
	return func(r *Runtime) { r.Config = cfg }
}

func WithRegistry(reg provider.ProviderRegistry) Option {
	return func(r *Runtime) { r.Registry = reg }
}

func WithManager(mgr tunnel.TunnelManager) Option {
	return func(r *Runtime) { r.Manager = mgr }
}

func New(opts ...Option) *Runtime {
	rt := &Runtime{
		Logger:   logger.NewDefault(),
		Registry: provider.NewRegistry(),
		Manager:  tunnel.NewManager(),
	}
	for _, opt := range opts {
		opt(rt)
	}
	return rt
}

func (r *Runtime) Shutdown(ctx context.Context) error {
	if r.Manager != nil {
		return r.Manager.StopAll(ctx)
	}
	return nil
}
