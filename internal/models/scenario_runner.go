package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	es "github.com/elastic/go-elasticsearch/v7"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/config"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/usage"
)

type OpType string

const (
	OpSearch OpType = "search"
	OpIndex         = "index"
)

type runningScenario struct {
	*Scenario

	exerciseCancelFunc   context.CancelFunc
	validationCancelFunc context.CancelFunc

	usageConn  *usage.Connection
	goldenConn *es.Client
}

type ScenarioRunner struct {
	cfg       *config.Config
	scenarios map[string]runningScenario
	usageConn *usage.Connection
}

func NewScenarioRunner(cfg *config.Config) (*ScenarioRunner, error) {
	sr := new(ScenarioRunner)
	sr.cfg = cfg
	sr.scenarios = map[string]runningScenario{}

	usageConn, err := sr.initUsageClusterConnection()
	if err != nil {
		return nil, err
	}

	sr.usageConn = usageConn
	return sr, nil
}

func (sr *ScenarioRunner) Start(s *Scenario) error {
	fmt.Println("starting scenario runner...")
	goldenConn, err := es.NewClient(es.Config{
		CloudID:  s.DeploymentCredentials.CloudID,
		Username: s.DeploymentCredentials.Username,
		Password: s.DeploymentCredentials.Password,
	})
	if err != nil {
		return fmt.Errorf("unable to create connection to golden deployment: %w", err)
	}

	if err := s.EnsureDeployment(sr.cfg); err != nil {
		return err
	}

	exerciseCtx, exerciseCancelFunc := context.WithCancel(context.Background())
	validationCtx, validationCancelFunc := context.WithCancel(context.Background())

	rs := runningScenario{
		Scenario:             s,
		exerciseCancelFunc:   exerciseCancelFunc,
		validationCancelFunc: validationCancelFunc,
		usageConn:            sr.usageConn,
		goldenConn:           goldenConn,
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

	startOffset := time.Duration(rs.Workload.StartOffsetSeconds) * time.Second
	now := time.Now()
	startTime := now.Add(startOffset)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("exercise loop done for scenario [%s]\n", rs.ID)
				ticker.Stop()
				return

			case t := <-ticker.C:
				if t.Before(startTime) {
					continue
				}

				numRequestsToFire := rand.Intn(rs.Workload.MaxRequestsPerSecond + 1)
				for i := 0; i < numRequestsToFire; i++ {
					op := randOp(rs.Workload.IndexToSearchRatio)
					switch op {
					case OpSearch:
						doSearch(rs.goldenConn, "foo*")

					case OpIndex:
						doIndex(rs.goldenConn, "foo", randIndexBody())
					}
				}
			}
		}
	}()
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

			case <-ticker.C:
				fmt.Printf("%s: running validations for scenario [%s]...\n", time.Now().Format(time.RFC3339), rs.ID)
				rs.Scenario.Validate(rs.usageConn)
			}
		}
	}()
}

func (sr *ScenarioRunner) initUsageClusterConnection() (*usage.Connection, error) {
	return usage.NewConnection(
		sr.cfg.UsageCluster.Url,
		sr.cfg.UsageCluster.Username,
		sr.cfg.UsageCluster.Password,
	)
}

func randOp(indexToSearchRatio int) OpType {
	ops := make([]OpType, 1+indexToSearchRatio)
	ops[0] = OpSearch
	for i := 1; i < len(ops); i++ {
		ops[i] = OpIndex
	}

	randIdx := rand.Intn(len(ops))
	return ops[randIdx]
}

func randIndexBody() json.RawMessage {
	messages := []string{
		"the quick brown fox",
		"jumped over the",
		"lazy dog",
	}

	innerKeys := []string{"count", "sum"}

	randMsgIdx := rand.Intn(len(messages))
	randMsg := messages[randMsgIdx]

	randKeyIdx := rand.Intn(len(innerKeys))
	randKey := innerKeys[randKeyIdx]

	randNum := (17 + rand.Intn(10000)) % 523

	bodyTpl := `{"message":"%s","metric":{"%s":%d}}`
	body := fmt.Sprintf(bodyTpl, randMsg, randKey, randNum)

	return json.RawMessage(body)
}

func doSearch(esClient *es.Client, target string) error {
	if _, err := esClient.Search(
		esClient.Search.WithIndex(target),
	); err != nil {
		return fmt.Errorf("search operation failed: %w", err)
	}

	return nil
}

func doIndex(esClient *es.Client, target string, body json.RawMessage) error {
	var b bytes.Buffer
	if len(body) > 0 {
		b.Write(body)
	}

	_, err := esClient.Index(
		target,
		&b,
	)

	if err != nil {
		return fmt.Errorf("index operation failed: %w", err)
	}

	return nil
}
