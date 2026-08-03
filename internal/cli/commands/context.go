package commands

import (
	"context"

	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/alexperezortuno/portx/internal/tunnel"
)

type RegistryKey struct{}
type ManagerKey struct{}
type ConfigKey struct{}

func GetRegistry(ctx context.Context) provider.ProviderRegistry {
	if v := ctx.Value(RegistryKey{}); v != nil {
		return v.(provider.ProviderRegistry)
	}
	return provider.NewRegistry()
}

func GetManager(ctx context.Context) tunnel.TunnelManager {
	if v := ctx.Value(ManagerKey{}); v != nil {
		return v.(tunnel.TunnelManager)
	}
	return tunnel.NewManager()
}
