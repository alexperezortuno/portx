package commands

import (
	"fmt"
	"log/slog"

	"github.com/alexperezortuno/portx/internal/config"
	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/alexperezortuno/portx/internal/provider/cloudflare"
	"github.com/alexperezortuno/portx/internal/provider/portxd"
	"github.com/alexperezortuno/portx/internal/provider/ssh"
	"github.com/alexperezortuno/portx/internal/sshutil"
	"github.com/alexperezortuno/portx/internal/target"
	"github.com/spf13/cobra"
)

type ExposeOptions struct {
	Target      string
	Provider    string
	LocalAddr   string
	RemoteAddr  string
	Hostname    string
	SSHUser     string
	SSHHost     string
	SSHPort     int
	SSHPassword string
	SSHPKey     string
	SSHUseAgent bool
	PortXDPort  int
	PortFlag    int
}

func NewExposeCommand() *cobra.Command {
	opts := &ExposeOptions{}
	cmd := &cobra.Command{
		Use:   "expose [target]",
		Short: "Expose a local service through a tunnel",
		Long: `Expose a local service to the internet through a tunnel provider.

Target can be:
  - A port number: "portx expose 3000"
  - A named service from config: "portx expose frontend"

Examples:
  # Expose local port 3000 (uses default provider: portxd)
  portx expose 3000

  # Expose named service from config
  portx expose frontend

  # Expose with specific provider
  portx expose 3000 --provider ssh --ssh-host example.com --ssh-use-agent

  # Expose via Cloudflare Quick Tunnel (no account needed)
  portx expose 8080 --provider cloudflare

  # Override service config
  portx expose frontend --provider cloudflare`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Target = args[0]
			}
			return runExpose(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Tunnel provider (ssh, portxd, cloudflare)")
	cmd.Flags().IntVar(&opts.PortFlag, "port", 0, "Local port (deprecated: use positional argument)")
	cmd.Flags().StringVar(&opts.LocalAddr, "local-addr", "", "Local service address (host:port)")
	cmd.Flags().StringVar(&opts.RemoteAddr, "tunnel-port", "", "Tunnel address on remote (host:port)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Hostname for the tunnel (e.g., api.example.com)")
	cmd.Flags().StringVar(&opts.SSHUser, "ssh-user", "", "SSH username")
	cmd.Flags().StringVar(&opts.SSHHost, "ssh-host", "", "SSH server host")
	cmd.Flags().IntVar(&opts.SSHPort, "ssh-port", 22, "SSH server port")
	cmd.Flags().StringVar(&opts.SSHPassword, "ssh-password", "", "SSH password")
	cmd.Flags().StringVar(&opts.SSHPKey, "ssh-private-key", "", "SSH private key content (PEM format)")
	cmd.Flags().BoolVar(&opts.SSHUseAgent, "ssh-use-agent", false, "Use SSH agent for authentication")
	cmd.Flags().IntVar(&opts.PortXDPort, "portxd-port", 7222, "PortXD server port")

	return cmd
}

func runExpose(cmd *cobra.Command, opts *ExposeOptions) error {
	ctx := cmd.Context()
	logger := slog.Default()

	cfg := GetConfig(ctx)
	if cfg == nil {
		cfg = &config.Config{}
	}

	targetStr := opts.Target
	if targetStr == "" && opts.PortFlag > 0 {
		targetStr = fmt.Sprintf("%d", opts.PortFlag)
	}

	if targetStr == "" {
		return fmt.Errorf("target is required (port number or service name)")
	}

	resolver := target.NewResolver(cfg)
	resolved, err := resolver.Resolve(targetStr)
	if err != nil {
		return fmt.Errorf("resolving target %q: %w", targetStr, err)
	}

	providerName := opts.Provider
	if providerName == "" {
		providerName = cfg.DefaultProvider()
	}

	localAddr := opts.LocalAddr
	if localAddr == "" {
		localAddr = resolved.LocalAddr
	}

	registry := GetRegistry(ctx)
	manager := GetManager(ctx)

	var p provider.Provider

	switch providerName {
	case "portxd":
		p = portxd.New("portxd", portxd.ProviderConfig{
			LocalAddr:  localAddr,
			ServerPort: opts.PortXDPort,
		})

	case "ssh":
		p = buildSSHProvider(opts)

	case "cloudflare":
		p = cloudflare.New("cloudflare")

	default:
		return fmt.Errorf("unsupported provider: %q (supported: ssh, portxd, cloudflare)", providerName)
	}

	if err := registry.Register(p); err != nil {
		return fmt.Errorf("registering provider: %w", err)
	}

	tunnelName := fmt.Sprintf("%s-%d", providerName, resolved.Port)
	if resolved.Type == target.TypeService {
		tunnelName = fmt.Sprintf("%s-%s", providerName, resolved.Name)
	}

	tunnelCfg := provider.TunnelConfig{
		LocalAddr:  localAddr,
		RemoteAddr: opts.RemoteAddr,
	}

	if err := manager.Add(ctx, tunnelName, tunnelCfg, p); err != nil {
		return fmt.Errorf("adding tunnel: %w", err)
	}

	if err := manager.StartAll(ctx); err != nil {
		return fmt.Errorf("starting tunnel: %w", err)
	}

	logger.Info("Tunnel started",
		"name", tunnelName,
		"provider", providerName,
		"target", targetStr,
		"local", localAddr,
	)

	<-cmd.Context().Done()
	return nil
}

func buildSSHProvider(opts *ExposeOptions) provider.Provider {
	cfg := sshutil.Config{
		User:       opts.SSHUser,
		Host:       opts.SSHHost,
		Port:       opts.SSHPort,
		Password:   opts.SSHPassword,
		PrivateKey: opts.SSHPKey,
		RemoteAddr: opts.RemoteAddr,
		UseAgent:   opts.SSHUseAgent,
	}
	return ssh.New("ssh", cfg)
}
