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
	// TODO: add flags for test results cluster

	rootCmd.AddCommand(setUpCmd)
	//rootCmd.AddCommand(exerciseCmd)
	//rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(tearDownCmd)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("could not execute root command: %w", err)
	}

	return nil
}
