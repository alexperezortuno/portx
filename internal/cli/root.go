package cli

import (
	"context"
	"log/slog"

	"github.com/alexperezortuno/portx/internal/cli/commands"
	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/alexperezortuno/portx/internal/tunnel"
	"github.com/spf13/cobra"
)

type contextKey string

type RootCommand struct {
	cmd      *cobra.Command
	logger   *slog.Logger
	registry provider.ProviderRegistry
	manager  tunnel.TunnelManager
}

func NewRoot() *RootCommand {
	root := &cobra.Command{
		Use:           "portx",
		Short:         "Provider agnostic tunneling platform",
		Long:          "PortX is a provider-agnostic tunneling platform that supports multiple providers including SSH, Cloudflare, FRP, and more.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rc := &RootCommand{
		cmd:      root,
		logger:   slog.Default(),
		registry: provider.NewRegistry(),
		manager:  tunnel.NewManager(),
	}

	root.AddCommand(commands.NewVersionCommand())
	root.AddCommand(commands.NewDoctorCommand())
	root.AddCommand(commands.NewExposeCommand())

	return rc
}

func (r *RootCommand) Execute(args []string) error {
	r.cmd.SetArgs(args[1:])

	ctx := context.Background()
	ctx = context.WithValue(ctx, contextKey("registry"), r.registry)
	ctx = context.WithValue(ctx, contextKey("manager"), r.manager)

	return r.cmd.ExecuteContext(ctx)
}

func GetRegistry(cmd *cobra.Command) provider.ProviderRegistry {
	if v := cmd.Context().Value(contextKey("registry")); v != nil {
		return v.(provider.ProviderRegistry)
	}
	return provider.NewRegistry()
}

func GetManager(cmd *cobra.Command) tunnel.TunnelManager {
	if v := cmd.Context().Value(contextKey("manager")); v != nil {
		return v.(tunnel.TunnelManager)
	}
	return tunnel.NewManager()
}
