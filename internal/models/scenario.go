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

type IntRange struct {
	Min int `json:"min" binding:"required"`
	Max int `json:"max" binding:"required"`
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

	ID        string    `json:"id"`
	StartedOn time.Time `json:"started_on"`
	StoppedOn time.Time `json:"stopped_on"`
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

	s.StartedOn = time.Now()
	return s.persist()
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
