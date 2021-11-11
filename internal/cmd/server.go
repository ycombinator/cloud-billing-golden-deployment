package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Starting API server...")

		if _, exists := os.LookupEnv("EC_API_KEY"); !exists {
			return fmt.Errorf("Elastic Cloud API KEY environment variable [EC_API_KEY] is not set")
		}

		if err := server.Start(); err != nil {
			return err
		}

		return nil
	},
}
