package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"

	"github.com/spf13/cobra"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Scenario Runner and API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateServerCmdInput(); err != nil {
			return err
		}

		setupCloseHandler()

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

func validateServerCmdInput() error {
	if _, exists := os.LookupEnv("EC_GOLDEN_API_KEY"); !exists {
		return fmt.Errorf("Elastic Cloud Golden Deployment Account API Key environment variable [EC_GOLDEN_API_KEY] is not set")
	}

	if _, exists := os.LookupEnv("EC_USAGE_URL"); !exists {
		return fmt.Errorf("Elastic Cloud Usage Cluster URL environment variable [EC_USAGE_URL] is not set")
	}

	if _, exists := os.LookupEnv("EC_USAGE_API_KEY"); !exists {
		return fmt.Errorf("Elasticsearch Usage Cluster API Key environment variable [EC_USAGE_API_KEY] is not set")
	}

	return nil
}

func initScenarioRunner() error {
	// Load all scenarios
	scenarios, err := models.LoadAllScenarios()
	if err != nil {
		return fmt.Errorf("could not load all scenarios: %w", err)
	}

	// Get scenario runner singleton
	scenarioRunner, err := models.NewScenarioRunner()
	if err != nil {
		return err
	}

	// Ask scenario runner to run each scenario that's started
	for _, scenario := range scenarios {
		scenarioRunner.Start(&scenario)
	}

	return nil
}

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Printf("Stopping Scenario Runner... ")

		scenarioRunner, err := models.NewScenarioRunner()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		scenarioRunner.StopAll()
		fmt.Println("done")

		os.Exit(0)
	}()
}
