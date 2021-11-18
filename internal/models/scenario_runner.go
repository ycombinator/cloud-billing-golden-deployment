package models

import (
	"context"
	"fmt"
	"time"
)

var (
	scenarioRunnerSingleton *ScenarioRunner
)

type runningScenario struct {
	Scenario

	exerciseCancelFunc   context.CancelFunc
	validationCancelFunc context.CancelFunc
}

type ScenarioRunner struct {
	scenarios map[string]runningScenario
}

func NewScenarioRunnerSingleton() *ScenarioRunner {
	if scenarioRunnerSingleton == nil {
		scenarioRunnerSingleton = new(ScenarioRunner)
		scenarioRunnerSingleton.scenarios = map[string]runningScenario{}
	}

	return scenarioRunnerSingleton
}

func (sr *ScenarioRunner) Start(s Scenario) {
	fmt.Println("starting scenario runner...")
	exerciseCtx, exerciseCancelFunc := context.WithCancel(context.Background())
	validationCtx, validationCancelFunc := context.WithCancel(context.Background())

	rs := runningScenario{
		Scenario:             s,
		exerciseCancelFunc:   exerciseCancelFunc,
		validationCancelFunc: validationCancelFunc,
	}

	sr.scenarios[s.ID] = rs
	rs.start(exerciseCtx, validationCtx)
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
				// TODO: run validations
				fmt.Printf("%s: running validations for scenario [%s]...\n", time.Now().Format(time.RFC3339), rs.ID)
			}
		}
	}()
}
