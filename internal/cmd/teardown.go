package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/terraform"
)

var tearDownCmd = &cobra.Command{
	Use:   "teardown",
	Short: "Tear down a test scenario",
	RunE: func(cmd *cobra.Command, args []string) error {
		scenario, err := cmd.Flags().GetString("scenario")
		if err != nil {
			return err
		}
		fmt.Printf("Tearing down scenario [%s]...\n", scenario)

		if _, exists := os.LookupEnv("EC_API_KEY"); !exists {
			return fmt.Errorf("Elastic Cloud API KEY environment variable [EC_API_KEY] is not set")
		}

		workDir, err := terraform.NewWorkDir(filepath.Join("deployment_configs", scenario, "setup"))
		if err != nil {
			return fmt.Errorf("could not load scenario [%s]: %w", scenario, err)
		}

		if err := workDir.Destroy(); err != nil {
			return fmt.Errorf("could not tear down scenario [%s]: %w", scenario, err)
		}

		fmt.Println("Done")
		return nil
	},
}

func init() {
	tearDownCmd.Flags().StringP("scenario", "s", "", "name of scenario")
	tearDownCmd.MarkFlagRequired("scenario")
}
