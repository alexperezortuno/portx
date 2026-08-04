package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alexperezortuno/portx/internal/tunnel"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	var outputFormat string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active tunnels",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, outputFormat)
		},
	}
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json")
	return cmd
}

func runList(cmd *cobra.Command, format string) error {
	mgr := GetManager(cmd.Context())
	tunnels := mgr.List(cmd.Context())

	switch strings.ToLower(format) {
	case "json":
		return outputJSON(tunnels)
	case "table":
		return outputTable(tunnels)
	default:
		return fmt.Errorf("unknown output format: %s (use table or json)", format)
	}
}

func outputTable(tunnels []*tunnel.Tunnel) error {
	if len(tunnels) == 0 {
		fmt.Println("No active tunnels")
		return nil
	}

	fmt.Printf("%-20s %-12s %-22s %-22s %s\n", "NAME", "PROVIDER", "LOCAL", "REMOTE", "STATUS")
	fmt.Println(strings.Repeat("-", 100))

	for _, t := range tunnels {
		local := t.Config.LocalAddr
		if local == "" {
			local = "-"
		}
		remote := t.Config.RemoteAddr
		if remote == "" {
			if u, ok := t.Provider.(interface{ URL() string }); ok {
				remote = u.URL()
			}
		}
		if remote == "" {
			remote = "-"
		}
		fmt.Printf("%-20s %-12s %-22s %-22s %s\n", t.Name, t.Provider.Name(), local, remote, t.Status)
	}
	return nil
}

func outputJSON(tunnels []*tunnel.Tunnel) error {
	type tunnelOut struct {
		Name     string            `json:"name"`
		Provider string            `json:"provider"`
		Config   map[string]string `json:"config"`
		Status   string            `json:"status"`
		URL      string            `json:"url,omitempty"`
	}

	out := make([]tunnelOut, len(tunnels))
	for i, t := range tunnels {
		tun := tunnelOut{
			Name:     t.Name,
			Provider: t.Provider.Name(),
			Config: map[string]string{
				"local_addr":  t.Config.LocalAddr,
				"remote_addr": t.Config.RemoteAddr,
			},
			Status: string(t.Status),
		}
		if u, ok := t.Provider.(interface{ URL() string }); ok {
			tun.URL = u.URL()
		}
		out[i] = tun
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
