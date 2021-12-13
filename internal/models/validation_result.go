package models

import "time"

type ValidationResult struct {
	ScenarioID string `json:"scenario_id"`

	ValidatedOn time.Time `json:"validated_on"`

	InstanceCapacityGBHours  FloatValidationResult `json:"instance_capacity_gb_hours"`
	DataOutGB                FloatValidationResult `json:"data_out_gb"`
	DataInterNodeGB          FloatValidationResult `json:"data_internode_gb"`
	SnapshotStorageSizeGB    FloatValidationResult `json:"snapshot_storage_size_gb"`
	SnapshotAPIRequestsCount FloatValidationResult `json:"snapshot_api_requests_count"`
}
