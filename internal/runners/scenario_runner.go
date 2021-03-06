package runners

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/logging"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/deployment"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/config"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/dao"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/usage"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/auth"

	es "github.com/elastic/go-elasticsearch/v7"
)

type OpType string

const (
	OpSearch OpType = "search"
	OpIndex         = "index"
)

type runningScenario struct {
	*models.Scenario

	exerciseCancelFunc   context.CancelFunc
	validationCancelFunc context.CancelFunc

	usageConn  *usage.Connection
	stateConn  *es.Client
	goldenConn *es.Client
}

type ScenarioRunner struct {
	cfg       *config.Config
	scenarios map[string]runningScenario

	usageConn *usage.Connection
	stateConn *es.Client
	essConn   *api.API
}

func NewScenarioRunner(cfg *config.Config) (*ScenarioRunner, error) {
	sr := new(ScenarioRunner)
	sr.cfg = cfg
	sr.scenarios = map[string]runningScenario{}

	usageConn, err := sr.initUsageClusterConnection()
	if err != nil {
		return nil, err
	}

	stateConn, err := es.NewClient(es.Config{
		Addresses: []string{cfg.StateCluster.Url},
		Username:  cfg.StateCluster.Username,
		Password:  cfg.StateCluster.Password,
	})
	if err != nil {
		return nil, err
	}

	essConn, err := api.NewAPI(api.Config{
		Host:       cfg.API.Url,
		Client:     new(http.Client),
		AuthWriter: auth.APIKey(cfg.API.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Elastic Cloud API at [%s]: %w", cfg.API.Url, err)
	}

	sr.usageConn = usageConn
	sr.stateConn = stateConn
	sr.essConn = essConn

	return sr, nil
}

func (sr *ScenarioRunner) Start(s *models.Scenario) error {
	logging.Logger.Info("starting scenario runner...")

	deploymentName := s.GetDeploymentName()
	exists, err := deployment.CheckIfDeploymentExists(sr.essConn, deploymentName)
	if err != nil {
		return fmt.Errorf("unable to check if deployment [%s] exists: %w", deploymentName, err)
	}

	if !exists {
		// Create deployment
		deploymentConfigDAO := dao.NewDeploymentConfiguration(sr.stateConn)
		deploymentConfig, err := deploymentConfigDAO.Get(s.DeploymentConfiguration.ID)
		if err != nil {
			return err
		}

		if deploymentConfig == nil {
			return fmt.Errorf("deployment configuration [%s] specified in scenario [%s] does not exist", s.DeploymentConfiguration.ID, s.ID)
		}

		req, err := deploymentConfig.ToDeploymentCreateRequest(s.DeploymentConfiguration.Variables)
		if err != nil {
			return fmt.Errorf("unable to create deployment create request from configuration [%s]: %w", deploymentConfig.ID, err)
		}

		logging.Logger.Info("creating deployment from config...",
			zap.String("deployment", deploymentName),
			zap.String("config", deploymentConfig.ID),
		)
		out, err := deployment.CreateDeployment(sr.essConn, deploymentName, req)
		if err != nil {
			return err
		}

		s.ClusterIDs = out.ClusterIDs
		s.DeploymentCredentials = out.DeploymentCredentials
	}

	goldenConn, err := es.NewClient(es.Config{
		CloudID:  s.DeploymentCredentials.CloudID,
		Username: s.DeploymentCredentials.Username,
		Password: s.DeploymentCredentials.Password,
	})
	if err != nil {
		return fmt.Errorf("unable to create connection to golden deployment: %w", err)
	}

	exerciseCtx, exerciseCancelFunc := context.WithCancel(context.Background())
	validationCtx, validationCancelFunc := context.WithCancel(context.Background())

	rs := runningScenario{
		Scenario:             s,
		exerciseCancelFunc:   exerciseCancelFunc,
		validationCancelFunc: validationCancelFunc,
		usageConn:            sr.usageConn,
		stateConn:            sr.stateConn,
		goldenConn:           goldenConn,
	}

	sr.scenarios[s.ID] = rs

	if s.StartedOn == nil {
		scenarioDAO := dao.NewScenario(sr.stateConn)
		if err := scenarioDAO.Save(s); err != nil {
			return err
		}

		now := time.Now()
		s.StartedOn = &now
	}

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
	loggingParam := zap.String("scenario", rs.ID)
	logging.Logger.Info("starting exercise loop", loggingParam)

	// TODO: consider start time + offset, if it has already passed, start now!
	startOffset := time.Duration(rs.Workload.StartOffsetSeconds) * time.Second
	now := time.Now()
	startTime := now.Add(startOffset)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				logging.Logger.Info("exercise loop done for scenario", loggingParam)
				ticker.Stop()
				return

			case t := <-ticker.C:
				if t.Before(startTime) {
					logging.Logger.Debug("not yet time to start exercising scenario", loggingParam)
					continue
				}

				numRequestsToFire := rand.Intn(rs.Workload.MaxRequestsPerSecond + 1)
				logging.Logger.Debug("firing requests now...", loggingParam, zap.Int("requests", numRequestsToFire))
				var err error
				for i := 0; i < numRequestsToFire; i++ {
					op := randOp(rs.Workload.IndexToSearchRatio)
					switch op {
					case OpSearch:
						logging.Logger.Debug("firing search request", loggingParam)
						err = doSearch(rs.goldenConn, "foo*")

					case OpIndex:
						logging.Logger.Debug("firing index request", loggingParam)
						err = doIndex(rs.goldenConn, "foo", randIndexBody())
					}
				}

				if err != nil {
					logging.Logger.Error(err.Error(), loggingParam)
				}
			}
		}
	}()
}

func (rs *runningScenario) startValidationLoop(ctx context.Context) {
	validationFrequency := rs.GetValidationFrequency()
	startAfter := waitFor(*rs.StartedOn, validationFrequency)

	loggingParam := zap.String("scenario", rs.ID)
	logging.Logger.Info("starting validation loop", loggingParam, zap.Duration("delay", startAfter))

	var timer *time.Timer
	timer = time.AfterFunc(startAfter, func() {
		rs.validate()

		ticker := time.NewTicker(validationFrequency)
		for {
			select {
			case <-ctx.Done():
				logging.Logger.Info("stopping validation loop", loggingParam)
				ticker.Stop()
				timer.Stop()
				return

			case <-ticker.C:
				rs.validate()
			}
		}
	})
}

func (rs *runningScenario) validate() {
	loggingParam := zap.String("scenario", rs.ID)
	logging.Logger.Info("running validations...", loggingParam)
	result := rs.Scenario.Validate(rs.usageConn)

	validationResultDAO := dao.NewValidationResult(rs.stateConn)
	if err := validationResultDAO.Save(result); err != nil {
		logging.Logger.Error("error saving validation result", loggingParam, zap.Error(err))
	}
}

func waitFor(start time.Time, interval time.Duration) time.Duration {
	next := start
	for next.Before(time.Now()) {
		next = next.Add(interval)
	}

	return next.Sub(time.Now())
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
