package models

import (
	"encoding/json"
	"fmt"

	"github.com/cbroglie/mustache"
	"github.com/elastic/cloud-sdk-go/pkg/models"
)

type DeploymentConfiguration struct {
	ID   string `json:"id"`
	Vars map[string]struct {
		Type    string      `json:"type"`
		Default interface{} `json:"default"`
	} `json:"vars"`
	Template json.RawMessage `json:"template"`
}

func (dt *DeploymentConfiguration) ToDeploymentCreateRequest(overrideVars map[string]interface{}) (*models.DeploymentCreateRequest, error) {
	vars, err := dt.computeVars(overrideVars)
	if err != nil {
		return nil, err
	}

	contents, err := json.Marshal(dt)
	if err != nil {
		return nil, err
	}

	ctxt := map[string]interface{}{
		"vars": vars,
	}
	tplStr, err := mustache.Render(string(contents), ctxt)
	if err != nil {
		return nil, fmt.Errorf("unable to read deployment configuration for configuration [%s]: %w", dt.ID, err)
	}

	var tpl struct {
		Template models.DeploymentCreateRequest `json:"template"`
	}

	if err := json.Unmarshal([]byte(tplStr), &tpl); err != nil {
		return nil, fmt.Errorf("unable to decode deployment configuration for configuration [%s]: %w", dt.ID, err)
	}

	return &tpl.Template, nil
}

func (dt *DeploymentConfiguration) computeVars(overrideVars map[string]interface{}) (map[string]interface{}, error) {
	// Validate
	for name := range overrideVars {
		_, exists := dt.Vars[name]
		if !exists {
			return nil, fmt.Errorf("undefined variable [%s] in configuration [%s]", name, dt.ID)
		}
	}

	// Override
	vars := make(map[string]interface{}, len(dt.Vars))
	for name, value := range dt.Vars {
		vars[name] = value.Default
		if v, exists := overrideVars[name]; exists {
			vars[name] = v
		}
	}

	return vars, nil
}
