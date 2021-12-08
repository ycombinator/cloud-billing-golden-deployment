package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	es "github.com/elastic/go-elasticsearch/v7"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/config"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"

	"github.com/spf13/cobra"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/server"
)

func init() {
	serverCmd.Flags().StringP("config-file", "c", "config/qa.yml", "path to config file")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Scenario Runner and API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgFilePath, err := cmd.Flags().GetString("config-file")
		if err != nil {
			return err
		}

		cfg, err := config.LoadFromFile(cfgFilePath)
		if err != nil {
			return err
		}

		// Get scenario runner
		scenarioRunner, err := models.NewScenarioRunner(cfg)
		if err != nil {
			return err
		}

		stateConn, err := es.NewClient(es.Config{
			Addresses: []string{cfg.StateCluster.Url},
			Username:  cfg.StateCluster.Username,
			Password:  cfg.StateCluster.Password,
		})
		if err != nil {
			return fmt.Errorf("unable to create connection to state cluster: %w", err)
		}

		setupCloseHandler(scenarioRunner, stateConn)

		fmt.Println("Starting existing scenarios...")
		if err := startScenarios(scenarioRunner, stateConn); err != nil {
			return err
		}

		fmt.Println("Starting API server...")
		if err := server.Start(scenarioRunner, stateConn); err != nil {
			return err
		}

		return nil
	},
}

func startScenarios(scenarioRunner *models.ScenarioRunner, stateConn *es.Client) error {
	// Load all scenarios
	scenarios, err := models.LoadAllScenarios(stateConn)
	if err != nil {
		return fmt.Errorf("could not load all scenarios: %w", err)
	}

	// Ask scenario runner to run each scenario that's started
	for _, scenario := range scenarios {
		scenarioRunner.Start(&scenario)
	}

	return nil
}

func setupCloseHandler(scenarioRunner *models.ScenarioRunner, stateConn *es.Client) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Printf("Stopping Scenario Runner... ")

		scenarioRunner.StopAll()
		fmt.Println("done")

		os.Exit(0)
	}()
}
