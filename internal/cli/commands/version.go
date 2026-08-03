package commands

import (
	"fmt"

	"github.com/alexperezortuno/portx/internal/version"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Get()
			fmt.Printf("PortX %s\n", info.Version)
			fmt.Printf("  commit: %s\n", info.Commit)
			fmt.Printf("  date: %s\n", info.Date)
			fmt.Printf("  go: %s\n", info.GoVersion)
			return nil
		},
	}
	return cmd
}
