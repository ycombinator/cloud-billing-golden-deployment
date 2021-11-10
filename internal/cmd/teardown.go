package cmd

import (
	"fmt"
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

		workDir, err := terraform.NewWorkDir(filepath.Join("scenarios", scenario))
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
