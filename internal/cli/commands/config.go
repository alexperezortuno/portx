package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexperezortuno/portx/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigView(cmd, "yaml")
		},
	}
	cmd.AddCommand(NewConfigViewCommand())
	cmd.AddCommand(NewConfigValidateCommand())
	return cmd
}

func NewConfigViewCommand() *cobra.Command {
	var outputFormat string
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View resolved configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigView(cmd, outputFormat)
		},
	}
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "yaml", "Output format: yaml, json")
	return cmd
}

func NewConfigValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate configuration file",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runConfigValidate,
	}
	return cmd
}

func runConfigView(cmd *cobra.Command, format string) error {
	cfg := GetConfig(cmd.Context())
	if cfg == nil {
		return fmt.Errorf("no configuration loaded")
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
	return nil
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	cfgPath := ""
	if len(args) > 0 {
		cfgPath = args[0]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Printf("Invalid: %v\n", err)
		return err
	}

	fmt.Println("Valid configuration")
	fmt.Printf("  Log level: %s\n", cfg.LogLevel)
	fmt.Printf("  Default provider: %s\n", cfg.Provider)
	fmt.Printf("  Tunnels: %d\n", len(cfg.Tunnels))
	for i, t := range cfg.Tunnels {
		fmt.Printf("    [%d] %s (provider=%s, local=%s)\n", i, t.Name, t.Provider, t.LocalAddr)
	}
	return nil
}

func GetConfig(ctx context.Context) *config.Config {
	if v := ctx.Value(ConfigKey{}); v != nil {
		return v.(*config.Config)
	}
	return nil
}
