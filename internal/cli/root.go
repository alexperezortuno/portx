package cli

import (
	"context"
	"log/slog"

	"github.com/alexperezortuno/portx/internal/cli/commands"
	"github.com/alexperezortuno/portx/internal/config"
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
	config   *config.Config
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
	root.AddCommand(commands.NewListCommand())
	root.AddCommand(commands.NewStopCommand())
	root.AddCommand(commands.NewConfigCommand())

	return rc
}

func (r *RootCommand) SetConfig(cfg *config.Config) {
	r.config = cfg
}

func (r *RootCommand) Execute(args []string) error {
	r.cmd.SetArgs(args[1:])

	ctx := context.Background()
	ctx = context.WithValue(ctx, commands.RegistryKey{}, r.registry)
	ctx = context.WithValue(ctx, commands.ManagerKey{}, r.manager)
	if r.config != nil {
		ctx = context.WithValue(ctx, commands.ConfigKey{}, r.config)
	}

	return r.cmd.ExecuteContext(ctx)
}

func (r *RootCommand) SetContext(ctx context.Context) {
	r.cmd.SetContext(ctx)
}
