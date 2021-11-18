package cmd

import (
	"fmt"
	"os"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"

	"github.com/spf13/cobra"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Scenario Runner and API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, exists := os.LookupEnv("EC_API_KEY"); !exists {
			return fmt.Errorf("Elastic Cloud API KEY environment variable [EC_API_KEY] is not set")
		}

		fmt.Println("Starting Scenario Runner...")
		if err := initScenarioRunner(); err != nil {
			return err
		}

		fmt.Println("Starting API server...")
		if err := server.Start(); err != nil {
			return err
		}

		return nil
	},
}

func initScenarioRunner() error {
	// Load all scenarios
	scenarios, err := models.LoadAllScenarios()
	if err != nil {
		return fmt.Errorf("could not load all scenarios: %w", err)
	}

	// Get scenario runner singleton
	scenarioRunner := models.NewScenarioRunnerSingleton()

	// Ask scenario runner to run each scenario that's started
	for _, scenario := range scenarios {
		scenarioRunner.Start(scenario)
	}

	return nil
}
