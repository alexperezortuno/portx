package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alexperezortuno/portx/internal/provider"
	"github.com/alexperezortuno/portx/internal/provider/portxd"
	"github.com/alexperezortuno/portx/internal/provider/ssh"
	"github.com/alexperezortuno/portx/internal/tunnel"
	"github.com/spf13/cobra"
)

type ExposeOptions struct {
	Provider    string
	LocalAddr   string
	RemoteAddr  string
	SSHUser     string
	SSHHost     string
	SSHPort     int
	SSHPassword string
	SSHPKey     string
	PortXDPort  int
}

func NewExposeCommand() *cobra.Command {
	opts := &ExposeOptions{}
	cmd := &cobra.Command{
		Use:   "expose",
		Short: "Expose a local service through a tunnel",
		Long: `Expose a local service to the internet through a tunnel provider.

Examples:
  portx expose --provider ssh --ssh-user tunnel --ssh-host example.com --local-port 8080
  portx expose --provider portxd --local-port 8080`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExpose(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Tunnel provider (ssh, portxd)")
	cmd.Flags().StringVar(&opts.LocalAddr, "local-port", "", "Local service address (host:port)")
	cmd.Flags().StringVar(&opts.RemoteAddr, "tunnel-port", "", "Tunnel address on remote (host:port)")
	cmd.Flags().StringVar(&opts.SSHUser, "ssh-user", "", "SSH username")
	cmd.Flags().StringVar(&opts.SSHHost, "ssh-host", "", "SSH server host")
	cmd.Flags().IntVar(&opts.SSHPort, "ssh-port", 22, "SSH server port")
	cmd.Flags().StringVar(&opts.SSHPassword, "ssh-password", "", "SSH password")
	cmd.Flags().StringVar(&opts.SSHPKey, "ssh-private-key", "", "SSH private key content")
	cmd.Flags().IntVar(&opts.PortXDPort, "portxd-port", 7222, "PortXD server port")

	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("local-port")

	return cmd
}

func runExpose(cmd *cobra.Command, opts *ExposeOptions) error {
	ctx := context.Background()
	logger := slog.Default()
	registry := provider.NewRegistry()
	manager := tunnel.NewManager()

	if !provider.IsKnown(opts.Provider) {
		return fmt.Errorf("unknown provider %q, known providers: %v", opts.Provider, provider.KnownProviders())
	}

	tunnelName := fmt.Sprintf("%s-tunnel", opts.Provider)

	localAddr := opts.LocalAddr
	if host, port, err := splitHostPort(opts.LocalAddr); err == nil && host == "" {
		localAddr = fmt.Sprintf("localhost:%s", port)
	}

	var p provider.Provider

	switch opts.Provider {
	case "ssh":
		if opts.SSHUser == "" || opts.SSHHost == "" {
			return fmt.Errorf("ssh-user and ssh-host are required for SSH provider")
		}
		cfg := ssh.SSHConfig{
			User:       opts.SSHUser,
			Host:       opts.SSHHost,
			Port:       opts.SSHPort,
			Password:   opts.SSHPassword,
			PrivateKey: opts.SSHPKey,
			RemoteAddr: opts.RemoteAddr,
			LocalAddr:  localAddr,
		}
		p = ssh.New("ssh", cfg)

	case "portxd":
		cfg := portxd.ProviderConfig{
			LocalAddr:  localAddr,
			ServerPort: opts.PortXDPort,
		}
		p = portxd.New("portxd", cfg)

	default:
		return fmt.Errorf("unsupported provider: %s", opts.Provider)
	}

	if err := registry.Register(p); err != nil {
		return fmt.Errorf("registering provider: %w", err)
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
		"provider", opts.Provider,
		"local", localAddr,
	)

	<-cmd.Context().Done()
	return nil
}

func splitHostPort(addr string) (host, port string, err error) {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i], addr[i+1:], nil
		}
	}
	return "", addr, nil
}
