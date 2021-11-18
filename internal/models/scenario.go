package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func ScenariosDir() string {
	return filepath.Join("data", "scenarios")
}

type IntRange struct {
	Min int `json:"min" binding:"required"`
	Max int `json:"max" binding:"required"`
}

type IntValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Actual   int      `json:"actual"`
	Expected IntRange `json:"expected"`
}

type Scenario struct {
	DeploymentConfiguration struct {
		ID        string                 `json:"id" binding:"required"`
		Variables map[string]interface{} `json:"variables"`
	} `json:"deployment_configuration" binding:"required"`
	Workload struct {
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
			InstanceCapacityGBHours  IntRange `json:"instance_capacity_gb_hours" binding:"required"`
			DataOutGB                IntRange `json:"data_out_gb" binding:"required"`
			DataInterNodeGB          IntRange `json:"data_internode_gb" binding:"required"`
			SnapshotStorageSizeGB    IntRange `json:"snapshot_storage_size_gb" binding:"required"`
			SnapshotAPIRequestsCount IntRange `json:"snapshot_api_requests_count" binding:"required"`
		} `json:"expectations"`
	} `json:"validations"`

	ID        string     `json:"id"`
	StartedOn *time.Time `json:"started_on,omitempty"`
	StoppedOn *time.Time `json:"stopped_on,omitempty"`

	ValidationResults []struct {
		ValidatedOn              time.Time           `json:"validated_on"`
		InstanceCapacityGBHours  IntValidationResult `json:"instance_capacity_gb_hours"`
		DataOutGB                IntValidationResult `json:"data_out_gb"`
		DataInterNodeGB          IntValidationResult `json:"data_internode_gb"`
		SnapshotStorageSizeGB    IntValidationResult `json:"snapshot_storage_size_gb"`
		SnapshotAPIRequestsCount IntValidationResult `json:"snapshot_api_requests_count"`
	} `json:"validation_results"`
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

func (s *Scenario) Validate() error {
	// TODO: perform validations
	return nil
}

func (s *Scenario) GenerateID() error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	s.ID = id.String()
	return s.persist()
}

func (s *Scenario) Start() error {
	if s.ID == "" {
		return fmt.Errorf("scenario does not have an ID")
	}

	now := time.Now()
	s.StartedOn = &now
	s.StoppedOn = nil

	// TODO: actually start scenario in goroutine
	// Get scenario runner singleton
	// Ask scenario runner to start running scenario

	return s.persist()
}

func (s *Scenario) GetValidationFrequency() time.Duration {
	// TODO: support frequencies other than "daily"
	//return 24 * time.Hour
	return 10 * time.Second
}

func (s *Scenario) persist() error {
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
