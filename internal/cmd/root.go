package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ecbgd",
	Short: "ecbgd is the Elastic Cloud Billing Golden Deployment CLI",
	Long: "The Elastic Cloud Billing Golden Deployment CLI manages golden " +
		"deployments for validating metering and billing implementations.",
}

func init() {
	// TODO: add flags for test results cluster? Use Usage Cluster itself?

	rootCmd.AddCommand(serverCmd)
	//rootCmd.AddCommand(setUpCmd) // TODO: remove command definition?
	//rootCmd.AddCommand(generateCmd) // TODO: remove command definition?
	//rootCmd.AddCommand(tearDownCmd) // TODO: remove command definition?
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("could not execute root command: %w", err)
	}

	return nil
}
