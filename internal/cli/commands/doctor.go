package commands

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
)

func NewDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system requirements",
		RunE:  runDoctor,
	}
	return cmd
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("PortX Doctor")

	fmt.Print("\nConnectivity:\n")

	hosts := []string{"ssh.github.com:443", "cloudflare.com:443"}
	for _, host := range hosts {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			fmt.Printf("  X %s: unreachable\n", host)
		} else {
			err := conn.Close()
			if err != nil {
				return err
			}
			fmt.Printf("  OK %s: reachable\n", host)
		}
	}

	fmt.Print("\nEnvironment:\n")
	dir, err := os.Getwd()
	if err == nil {
		fmt.Printf("  PWD: %s\n", dir)
	}

	return nil
}
