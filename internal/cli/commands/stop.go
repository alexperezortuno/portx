package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewStopCommand() *cobra.Command {
	var stopAll bool
	cmd := &cobra.Command{
		Use:   "stop [name]",
		Short: "Stop a tunnel or all tunnels",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(cmd, args, stopAll)
		},
	}
	cmd.Flags().BoolVar(&stopAll, "all", false, "Stop all tunnels")
	return cmd
}

func runStop(cmd *cobra.Command, args []string, stopAll bool) error {
	mgr := GetManager(cmd.Context())
	ctx := cmd.Context()

	if stopAll || len(args) == 0 {
		if err := mgr.StopAll(ctx); err != nil {
			return fmt.Errorf("stopping all tunnels: %w", err)
		}
		fmt.Println("All tunnels stopped")
		return nil
	}

	name := args[0]
	if err := mgr.Remove(ctx, name); err != nil {
		return fmt.Errorf("stopping tunnel %q: %w", name, err)
	}
	fmt.Printf("Tunnel %q stopped\n", name)
	return nil
}
