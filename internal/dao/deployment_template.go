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
	deploymentConfigsIndex = "gds-deployment-configs"
)

type DeploymentConfiguration struct {
	stateConn *es.Client
}

func NewDeploymentConfiguration(stateConn *es.Client) *DeploymentConfiguration {
	s := new(DeploymentConfiguration)
	s.stateConn = stateConn

	return s
}

func (dt *DeploymentConfiguration) ListAll() ([]models.DeploymentConfiguration, error) {
	var deploymentConfigs []models.DeploymentConfiguration

	if exists, err := dt.indexExists(); err != nil {
		return nil, err
	} else if !exists {
		return deploymentConfigs, nil
	}

	res, err := dt.stateConn.Search(
		dt.stateConn.Search.WithContext(context.Background()),
		dt.stateConn.Search.WithIndex(deploymentConfigsIndex),
		dt.stateConn.Search.WithSize(10000),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to list all deployment configurations: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, handleESAPIErrorResponse(res)
	}

	var r struct {
		Hits struct {
			Hits []struct {
				ID     string                         `json:"_id"`
				Source models.DeploymentConfiguration `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}

	for _, hit := range r.Hits.Hits {
		deploymentConfigs = append(deploymentConfigs, hit.Source)
	}
	return deploymentConfigs, nil
}

func (dt *DeploymentConfiguration) Get(id string) (*models.DeploymentConfiguration, error) {
	if exists, err := dt.indexExists(); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	res, err := dt.stateConn.Get(deploymentConfigsIndex, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get deployment configuration [%s]: %w", id, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, handleESAPIErrorResponse(res)
	}

	var r struct {
		ID     string                         `json:"_id"`
		Source models.DeploymentConfiguration `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}

	deploymentConfig := r.Source
	return &deploymentConfig, nil
}

func (dt *DeploymentConfiguration) Save(deploymentConfig *models.DeploymentConfiguration) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(deploymentConfig); err != nil {
		return fmt.Errorf("unable to encode deployment configuration [%s] as JSON: %w", deploymentConfig.ID, err)
	}

	res, err := dt.stateConn.Index(
		deploymentConfigsIndex,
		&buf,
		dt.stateConn.Index.WithDocumentID(deploymentConfig.ID),
	)
	if err != nil {
		return fmt.Errorf("unable to persist deployment configuration [%s]: %w", deploymentConfig.ID, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return handleESAPIErrorResponse(res)
	}

	return nil
}

func (dt *DeploymentConfiguration) indexExists() (bool, error) {
	res, err := dt.stateConn.Indices.Exists([]string{deploymentConfigsIndex})
	if err != nil {
		return false, fmt.Errorf("unable to check if deployment configurations exist: %w", err)
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
