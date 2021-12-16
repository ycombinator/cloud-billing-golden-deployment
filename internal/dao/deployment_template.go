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
	deploymentTemplatesIndex = "gds-deployment-configs"
)

type DeploymentTemplate struct {
	stateConn *es.Client
}

func NewDeploymentTemplate(stateConn *es.Client) *DeploymentTemplate {
	s := new(DeploymentTemplate)
	s.stateConn = stateConn

	return s
}

func (dt *DeploymentTemplate) ListAll() ([]models.DeploymentTemplate, error) {
	var deploymentTemplates []models.DeploymentTemplate

	if exists, err := dt.indexExists(); err != nil {
		return nil, err
	} else if !exists {
		return deploymentTemplates, nil
	}

	res, err := dt.stateConn.Search(
		dt.stateConn.Search.WithContext(context.Background()),
		dt.stateConn.Search.WithIndex(deploymentTemplatesIndex),
		dt.stateConn.Search.WithSize(10000),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to list all deployment templates: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, handleESAPIErrorResponse(res)
	}

	var r struct {
		Hits struct {
			Hits []struct {
				ID     string                    `json:"_id"`
				Source models.DeploymentTemplate `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}

	for _, hit := range r.Hits.Hits {
		deploymentTemplates = append(deploymentTemplates, hit.Source)
	}
	return deploymentTemplates, nil
}

func (dt *DeploymentTemplate) Get(id string) (*models.DeploymentTemplate, error) {
	if exists, err := dt.indexExists(); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	res, err := dt.stateConn.Get(deploymentTemplatesIndex, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get deployment template [%s]: %w", id, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, handleESAPIErrorResponse(res)
	}

	var r struct {
		ID     string                    `json:"_id"`
		Source models.DeploymentTemplate `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}

	deploymentTemplate := r.Source
	return &deploymentTemplate, nil
}

func (dt *DeploymentTemplate) Save(deploymentTemplate *models.DeploymentTemplate) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(deploymentTemplate); err != nil {
		return fmt.Errorf("unable to encode deployment template [%s] as JSON: %w", deploymentTemplate.ID, err)
	}

	res, err := dt.stateConn.Index(
		deploymentTemplatesIndex,
		&buf,
		dt.stateConn.Index.WithDocumentID(deploymentTemplate.ID),
	)
	if err != nil {
		return fmt.Errorf("unable to persist deployment template [%s]: %w", deploymentTemplate.ID, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return handleESAPIErrorResponse(res)
	}

	return nil
}

func (dt *DeploymentTemplate) indexExists() (bool, error) {
	res, err := dt.stateConn.Indices.Exists([]string{deploymentTemplatesIndex})
	if err != nil {
		return false, fmt.Errorf("unable to check if deployment templates exist: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return false, nil
		}
		return false, handleESAPIErrorResponse(res)
	}

	return true, nil
}
