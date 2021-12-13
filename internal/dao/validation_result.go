package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"
)

const (
	validationResultsIndex = "validation-results"
)

type ValidationResult struct {
	stateConn *es.Client
}

func NewValidationResult(stateConn *es.Client) *ValidationResult {
	vr := new(ValidationResult)
	vr.stateConn = stateConn

	return vr
}

func (vr *ValidationResult) ListAllForScenario(scenarioID string) ([]models.ValidationResult, error) {
	var results []models.ValidationResult

	res, err := vr.stateConn.Search(
		vr.stateConn.Search.WithContext(context.Background()),
		vr.stateConn.Search.WithIndex(validationResultsIndex),
		vr.stateConn.Search.WithSize(10000),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to list all validation results for scenario [%s]: %w", scenarioID, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, handleESAPIErrorResponse(res)
	}

	var r struct {
		Hits struct {
			Hits []struct {
				ID     string                  `json:"_id"`
				Source models.ValidationResult `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}

	for _, hit := range r.Hits.Hits {
		result := hit.Source
		results = append(results, result)
	}

	return results, nil
}

func (vr *ValidationResult) Save(result *models.ValidationResult) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(result); err != nil {
		return fmt.Errorf("unable to encode validation result for scenario [%s] as JSON: %w", result.ScenarioID, err)
	}

	res, err := vr.stateConn.Index(
		validationResultsIndex,
		&buf,
	)
	if err != nil {
		return fmt.Errorf("unable to persist validation result for scenario [%s]: %w", result.ScenarioID, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return handleESAPIErrorResponse(res)
	}

	return nil
}
