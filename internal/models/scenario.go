package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/config"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/deployment"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/usage"

	"github.com/google/uuid"
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

	Error error `json:"error"`
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
		StartOffsetSeconds int `json:"start_offset_seconds"`
		MinIntervalSeconds int `json:"min_interval_seconds"`
		MaxIntervalSeconds int `json:"max_interval_seconds"`
		IndexToSearchRatio int `json:"index_to_search_ratio"`
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

	ID        string     `json:"id"`
	ClusterID string     `json:"cluster_id"`
	StartedOn *time.Time `json:"started_on,omitempty"`
	StoppedOn *time.Time `json:"stopped_on,omitempty"`

	ValidationResults []ValidationResult `json:"validation_results"`
}

func LoadAllScenarios() ([]Scenario, error) {
	var scenarios []Scenario

	files, err := os.ReadDir(ScenariosDir())
	if err != nil {
		return nil, err
	}

	var dirnames []string
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		dirnames = append(dirnames, file.Name())
	}

	for _, dirname := range dirnames {
		scenario, err := LoadScenario(dirname)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, *scenario)
	}

	return scenarios, nil
}

func LoadScenario(id string) (*Scenario, error) {
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

func (s *Scenario) Validate(usageConn *usage.Connection) {
	q := usage.Query{
		ClusterID: s.ClusterID,
		From:      s.Validations.StartTimestamp,
		To:        s.Validations.EndTimestamp,
	}

	result := new(ValidationResult)
	result.ValidatedOn = time.Now()

	s.validateInstanceCapacity(usageConn, q, result)
	s.validateDataInterNode(usageConn, q, result)
	s.validateDataOut(usageConn, q, result)
	s.validateSnapshotAPIRequests(usageConn, q, result)
	s.validateSnapshotStorageSize(usageConn, q, result)

	s.ValidationResults = append(s.ValidationResults, *result)
	s.Persist()
}

func (s *Scenario) GenerateID() error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	s.ID = id.String()
	return s.Persist()
}

func (s *Scenario) EnsureDeployment(cfg *config.Config) error {
	if s.ClusterID != "" {
		return nil
	}

	out, err := deployment.EnsureDeployment(cfg, s.DeploymentTemplate)
	if err != nil {
		return err
	}

	s.ClusterID = out.ClusterID
	return s.Persist()
}

func (s *Scenario) Start(scenarioRunner *ScenarioRunner) error {
	if s.ID == "" {
		return fmt.Errorf("scenario does not have an ID")
	}

	now := time.Now()
	s.StartedOn = &now
	s.StoppedOn = nil

	if err := s.Persist(); err != nil {
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

func (s *Scenario) Persist() error {
	folder := filepath.Join("data", "scenarios", s.ID)
	_, err := os.Stat(folder)
	if err != nil {
		if !os.IsExist(err) {
			if err := os.Mkdir(folder, os.ModeDir|0755); err != nil {
				return fmt.Errorf("could not create scenario folder: %w", err)
			}
		} else {
			return fmt.Errorf("unexpected error reading scenario folder: %w", err)
		}
	}

	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("could not serialize scenario: %w", err)
	}

	file := filepath.Join(folder, "scenario.json")
	if err := ioutil.WriteFile(file, data, 0644); err != nil {
		return fmt.Errorf("could not persist scenario: %w", err)
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

func validateFloatRange(q usage.Query, f usage.FloatQueryFunc, expectations FloatRange, result *FloatValidationResult) {
	actual, err := f(q)
	if err != nil {
		result.Error = err
		return
	}

	result.Expected = expectations
	result.Actual = actual
	result.IsValid = expectations.IsInRange(actual)
}

func (ir *FloatRange) IsInRange(actual float64) bool {
	return ir.Min <= actual && actual <= ir.Max
}
