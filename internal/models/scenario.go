package models

import (
	"fmt"
	"time"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/deployment"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/usage"

	"github.com/google/uuid"
)

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

type Scenario struct {
	DeploymentTemplate struct {
		ID        string                 `json:"id" binding:"required"`
		Variables map[string]interface{} `json:"vars,omitempty"`
	} `json:"deployment_template" binding:"required"`
	Workload struct {
		StartOffsetSeconds   int `json:"start_offset_seconds"`
		MinIntervalSeconds   int `json:"min_interval_seconds"`
		MaxIntervalSeconds   int `json:"max_interval_seconds"`
		MaxRequestsPerSecond int `json:"max_requests_per_second"`
		IndexToSearchRatio   int `json:"index_to_search_ratio"`
	} `json:"workload"`
	Validations struct {
		FrequencySeconds int `json:"frequency_seconds"`
		Query            struct {
			StartTimestamp string `json:"start_timestamp"`
			EndTimestamp   string `json:"end_timestamp"`
		} `json:"query"`
		Expectations struct {
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
}

func (s *Scenario) IsStarted() bool {
	return s.StartedOn != nil && !s.StartedOn.IsZero()
}

func (s *Scenario) Validate(usageConn *usage.Connection) *ValidationResult {
	q := usage.Query{
		ClusterIDs: s.ClusterIDs,
		From:       s.Validations.Query.StartTimestamp,
		To:         s.Validations.Query.EndTimestamp,
	}

	result := new(ValidationResult)
	result.ScenarioID = s.ID
	result.ValidatedOn = time.Now()

	s.validateInstanceCapacity(usageConn, q, result)
	s.validateDataInterNode(usageConn, q, result)
	s.validateDataOut(usageConn, q, result)
	s.validateSnapshotAPIRequests(usageConn, q, result)
	s.validateSnapshotStorageSize(usageConn, q, result)

	return result
}

func (s *Scenario) GenerateID() error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	s.ID = id.String()
	return nil
}

func (s *Scenario) GetDeploymentName() string {
	return fmt.Sprintf("golden-%s", s.ID)
}

func (s *Scenario) GetValidationFrequency() time.Duration {
	return time.Duration(s.Validations.FrequencySeconds) * time.Second
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
