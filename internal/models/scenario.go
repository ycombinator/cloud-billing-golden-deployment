package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"

	es "github.com/elastic/go-elasticsearch/v7"

	"github.com/elastic/cloud-sdk-go/pkg/api"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/deployment"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/usage"

	"github.com/google/uuid"
)

const (
	ScenariosIndex = "scenarios"
)

func ScenariosDir() string {
	return filepath.Join("data", "scenarios")
}

type FloatRange struct {
	Min float64 `json:"min" binding:"required"`
	Max float64 `json:"max" binding:"required"`
}

type FloatValidationResult struct {
	IsValid bool `json:"is_valid"`

	Actual   float64    `json:"actual"`
	Expected FloatRange `json:"expected"`

	Error string `json:"error"`
}

type ValidationResult struct {
	ValidatedOn time.Time `json:"validated_on"`

	InstanceCapacityGBHours  FloatValidationResult `json:"instance_capacity_gb_hours"`
	DataOutGB                FloatValidationResult `json:"data_out_gb"`
	DataInterNodeGB          FloatValidationResult `json:"data_internode_gb"`
	SnapshotStorageSizeGB    FloatValidationResult `json:"snapshot_storage_size_gb"`
	SnapshotAPIRequestsCount FloatValidationResult `json:"snapshot_api_requests_count"`
}

type Scenario struct {
	DeploymentTemplate deployment.Template `json:"deployment_template" binding:"required"`
	Workload           struct {
		StartOffsetSeconds   int `json:"start_offset_seconds"`
		MinIntervalSeconds   int `json:"min_interval_seconds"`
		MaxIntervalSeconds   int `json:"max_interval_seconds"`
		MaxRequestsPerSecond int `json:"max_requests_per_second"`
		IndexToSearchRatio   int `json:"index_to_search_ratio"`
	} `json:"workload"`
	Validations struct {
		Frequency      string `json:"frequency"`
		StartTimestamp string `json:"start_timestamp"`
		EndTimestamp   string `json:"end_timestamp"`
		Expectations   struct {
			InstanceCapacityGBHours  FloatRange `json:"instance_capacity_gb_hours" binding:"required"`
			DataOutGB                FloatRange `json:"data_out_gb" binding:"required"`
			DataInterNodeGB          FloatRange `json:"data_internode_gb" binding:"required"`
			SnapshotStorageSizeGB    FloatRange `json:"snapshot_storage_size_gb" binding:"required"`
			SnapshotAPIRequestsCount FloatRange `json:"snapshot_api_requests_count" binding:"required"`
		} `json:"expectations"`
	} `json:"validations"`

	ID                    string                 `json:"id"`
	ClusterIDs            []string               `json:"cluster_ids"`
	DeploymentCredentials deployment.Credentials `json:"deployment_credentials"`

	StartedOn *time.Time `json:"started_on,omitempty"`
	StoppedOn *time.Time `json:"stopped_on,omitempty"`

	ValidationResults []ValidationResult `json:"validation_results"`
}

func LoadAllScenarios(stateConn *es.Client) ([]Scenario, error) {
	var scenarios []Scenario

	res, err := stateConn.Search(
		stateConn.Search.WithContext(context.Background()),
		stateConn.Search.WithIndex(ScenariosIndex),
		stateConn.Search.WithSize(10000),
		// TODO: add active scenarios filter
	)

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve all scenarios: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, handleESAPIErrorRespons(res)
	}

	var r struct {
		Hits struct {
			Hits []struct {
				ID     string   `json:"_id"`
				Source Scenario `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}

	for _, hit := range r.Hits.Hits {
		scenarios = append(scenarios, hit.Source)
	}

	return scenarios, nil
}

func LoadScenario(id string, stateConn *es.Client) (*Scenario, error) {
	fmt.Printf("loading scenario [%s] from disk...\n", id)
	path := filepath.Join(ScenariosDir(), id, "scenario.json")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var scenario Scenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		return nil, err
	}

	return &scenario, nil
}

func (s *Scenario) IsStarted() bool {
	return s.StartedOn != nil && !s.StartedOn.IsZero()
}

func (s *Scenario) Validate(usageConn *usage.Connection, stateConn *es.Client) {
	q := usage.Query{
		ClusterIDs: s.ClusterIDs,
		From:       s.Validations.StartTimestamp,
		To:         s.Validations.EndTimestamp,
	}

	result := new(ValidationResult)
	result.ValidatedOn = time.Now()

	s.validateInstanceCapacity(usageConn, q, result)
	s.validateDataInterNode(usageConn, q, result)
	s.validateDataOut(usageConn, q, result)
	s.validateSnapshotAPIRequests(usageConn, q, result)
	s.validateSnapshotStorageSize(usageConn, q, result)

	s.ValidationResults = append(s.ValidationResults, *result)
	s.Persist(stateConn)
}

func (s *Scenario) GenerateID(stateConn *es.Client) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	s.ID = id.String()
	return s.Persist(stateConn)
}

func (s *Scenario) EnsureDeployment(essConn *api.API, stateConn *es.Client) error {
	deploymentName := fmt.Sprintf("golden-%s", s.ID)

	if s.ClusterIDs != nil && len(s.ClusterIDs) > 0 {
		exists, err := deployment.CheckIfDeploymentExists(essConn, deploymentName)
		if err != nil {
			return fmt.Errorf("unable to check if deployment [%s] exists: %w", deploymentName, err)
		}

		if exists {
			return nil
		}
	}

	out, err := deployment.CreateDeployment(essConn, deploymentName, s.DeploymentTemplate)
	if err != nil {
		return err
	}

	s.ClusterIDs = out.ClusterIDs
	s.DeploymentCredentials = out.DeploymentCredentials
	return s.Persist(stateConn)
}

func (s *Scenario) Start(scenarioRunner *ScenarioRunner, stateConn *es.Client) error {
	if s.ID == "" {
		return fmt.Errorf("scenario does not have an ID")
	}

	now := time.Now()
	s.StartedOn = &now
	s.StoppedOn = nil

	if err := s.Persist(stateConn); err != nil {
		return err
	}

	if err := scenarioRunner.Start(s); err != nil {
		return err
	}

	return nil
}

func (s *Scenario) GetValidationFrequency() time.Duration {
	// TODO: support frequencies other than "daily"
	//return 24 * time.Hour
	return 10 * time.Second
}

func (s *Scenario) Persist(stateConn *es.Client) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(s); err != nil {
		return fmt.Errorf("unable to encode scenario [%s] as JSON: %w", s.ID, err)
	}

	res, err := stateConn.Index(
		ScenariosIndex,
		&buf,
		stateConn.Index.WithDocumentID(s.ID),
	)
	if err != nil {
		return fmt.Errorf("unable to persist scenario [%s]: %w", s.ID, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return handleESAPIErrorRespons(res)
	}

	return nil
}

func (s *Scenario) validateInstanceCapacity(usageConn *usage.Connection, q usage.Query, result *ValidationResult) {
	validateFloatRange(q, usageConn.GetInstanceCapacityGBHours, s.Validations.Expectations.InstanceCapacityGBHours, &result.InstanceCapacityGBHours)
}

func (s *Scenario) validateDataInterNode(usageConn *usage.Connection, q usage.Query, result *ValidationResult) {
	validateFloatRange(q, usageConn.GetDataInterNodeGB, s.Validations.Expectations.DataInterNodeGB, &result.DataInterNodeGB)
}

func (s *Scenario) validateDataOut(usageConn *usage.Connection, q usage.Query, result *ValidationResult) {
	validateFloatRange(q, usageConn.GetDataOutGB, s.Validations.Expectations.DataOutGB, &result.DataOutGB)
}

func (s *Scenario) validateSnapshotAPIRequests(usageConn *usage.Connection, q usage.Query, result *ValidationResult) {
	validateFloatRange(q, usageConn.GetSnapshotAPIRequestsCount, s.Validations.Expectations.SnapshotAPIRequestsCount, &result.SnapshotAPIRequestsCount)
}

func (s *Scenario) validateSnapshotStorageSize(usageConn *usage.Connection, q usage.Query, result *ValidationResult) {
	validateFloatRange(q, usageConn.GetSnapshotStorageSizeGB, s.Validations.Expectations.SnapshotStorageSizeGB, &result.SnapshotStorageSizeGB)
}

func validateFloatRange(q usage.Query, f func(usage.Query) (float64, error), expectations FloatRange, result *FloatValidationResult) {
	actual, err := f(q)
	if err != nil {
		result.Error = err.Error()
		return
	}

	result.Expected = expectations
	result.Actual = actual
	result.IsValid = expectations.IsInRange(actual)
}

func (ir *FloatRange) IsInRange(actual float64) bool {
	return ir.Min <= actual && actual <= ir.Max
}

func handleESAPIErrorRespons(res *esapi.Response) error {
	var e map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
		return fmt.Errorf("error parsing the response body: %w", err)
	} else {
		return fmt.Errorf("[%s] %s: %s",
			res.Status(),
			e["error"].(map[string]interface{})["type"],
			e["error"].(map[string]interface{})["reason"],
		)
	}
}
