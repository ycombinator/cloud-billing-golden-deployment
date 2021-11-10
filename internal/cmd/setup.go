package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/terraform"
)

var setUpCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up a test scenario",
	RunE: func(cmd *cobra.Command, args []string) error {
		scenario, err := cmd.Flags().GetString("scenario")
		if err != nil {
			return err
		}
		fmt.Printf("Setting up scenario [%s]...\n", scenario)

		workDir, err := terraform.NewWorkDir(filepath.Join("scenarios", scenario))
		if err != nil {
			return fmt.Errorf("could not load scenario [%s]: %w", scenario, err)
		}

		if err := workDir.Init(); err != nil {
			return fmt.Errorf("could not set up scenario [%s]: %w", scenario, err)
		}

		if err := workDir.Apply(); err != nil {
			return fmt.Errorf("could not set up scenario [%s]: %w", scenario, err)
		}

		fmt.Println("Done")
		return nil
	},
}

func init() {
	setUpCmd.Flags().StringP("scenario", "s", "", "name of scenario")
	setUpCmd.MarkFlagRequired("scenario")
}
