package models

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/config"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/usage"
)

var (
	scenarioRunnerSingleton *ScenarioRunner
)

type runningScenario struct {
	*Scenario

	exerciseCancelFunc   context.CancelFunc
	validationCancelFunc context.CancelFunc

	usageConn *usage.Connection
}

type ScenarioRunner struct {
	cfg       *config.Config
	scenarios map[string]runningScenario
	usageConn *usage.Connection
}

func NewScenarioRunner(cfg *config.Config) (*ScenarioRunner, error) {
	scenarioRunnerSingleton = new(ScenarioRunner)
	scenarioRunnerSingleton.cfg = cfg
	scenarioRunnerSingleton.scenarios = map[string]runningScenario{}

	usageConn, err := initUsageClusterConnection()
	if err != nil {
		return nil, err
	}

	scenarioRunnerSingleton.usageConn = usageConn
	return scenarioRunnerSingleton, nil
}

func (sr *ScenarioRunner) Start(s *Scenario) error {
	fmt.Println("starting scenario runner...")
	exerciseCtx, exerciseCancelFunc := context.WithCancel(context.Background())
	validationCtx, validationCancelFunc := context.WithCancel(context.Background())

	rs := runningScenario{
		Scenario:             s,
		exerciseCancelFunc:   exerciseCancelFunc,
		validationCancelFunc: validationCancelFunc,
		usageConn:            sr.usageConn,
	}

	if err := s.EnsureDeployment(sr.cfg); err != nil {
		return err
	}

	sr.scenarios[s.ID] = rs
	rs.start(exerciseCtx, validationCtx)

	return nil
}

func (sr *ScenarioRunner) Stop(scenarioID string) {
	rs := sr.scenarios[scenarioID]

	rs.validationCancelFunc()
	rs.exerciseCancelFunc()

	delete(sr.scenarios, scenarioID)
}

func (sr *ScenarioRunner) StopAll() {
	for _, scenario := range sr.scenarios {
		sr.Stop(scenario.ID)
	}
}

func (rs *runningScenario) start(exerciseCtx, validationCtx context.Context) {
	rs.startExerciseLoop(exerciseCtx)
	rs.startValidationLoop(validationCtx)
}

func (rs *runningScenario) startExerciseLoop(ctx context.Context) {
	fmt.Println("starting exercise loop...")
}

func (rs *runningScenario) startValidationLoop(ctx context.Context) {
	fmt.Printf("starting validation loop for scenario [%s]...\n", rs.ID)
	validationFrequency := rs.GetValidationFrequency()
	ticker := time.NewTicker(validationFrequency)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("validation loop done for scenario [%s]\n", rs.ID)
				ticker.Stop()
				return

			case _ = <-ticker.C:
				fmt.Printf("%s: running validations for scenario [%s]...\n", time.Now().Format(time.RFC3339), rs.ID)
				rs.Scenario.Validate(rs.usageConn)
			}
		}
	}()
}

func initUsageClusterConnection() (*usage.Connection, error) {
	address := os.Getenv("EC_USAGE_URL")
	apiKey := os.Getenv("EC_USAGE_API_KEY")

	return usage.NewConnection(address, apiKey)
}
