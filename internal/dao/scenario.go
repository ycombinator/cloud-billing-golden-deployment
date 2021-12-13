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
	scenariosIndex = "scenarios"
)

type Scenario struct {
	stateConn *es.Client
}

func NewScenario(stateConn *es.Client) *Scenario {
	s := new(Scenario)
	s.stateConn = stateConn

	return s
}

func (s *Scenario) ListAll() ([]models.Scenario, error) {
	var scenarios []models.Scenario

	res, err := s.stateConn.Search(
		s.stateConn.Search.WithContext(context.Background()),
		s.stateConn.Search.WithIndex(scenariosIndex),
		s.stateConn.Search.WithSize(10000),
		// TODO: add active scenarios filter
	)

	if err != nil {
		return nil, fmt.Errorf("unable to list all scenarios: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, handleESAPIErrorResponse(res)
	}

	var r struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source models.Scenario `json:"_source"`
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

func (s *Scenario) Get(id string) (*models.Scenario, error) {
	res, err := s.stateConn.Get(scenariosIndex, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get scenario [%s]: %w", id, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, handleESAPIErrorResponse(res)
	}

	var r struct {
		ID     string          `json:"_id"`
		Source models.Scenario `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}

	scenario := r.Source
	return &scenario, nil
}

func (s *Scenario) Save(scenario *models.Scenario) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(s); err != nil {
		return fmt.Errorf("unable to encode scenario [%s] as JSON: %w", scenario.ID, err)
	}

	res, err := s.stateConn.Index(
		scenariosIndex,
		&buf,
		s.stateConn.Index.WithDocumentID(scenario.ID),
	)
	if err != nil {
		return fmt.Errorf("unable to persist scenario [%s]: %w", scenario.ID, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return handleESAPIErrorResponse(res)
	}

	return nil
}
