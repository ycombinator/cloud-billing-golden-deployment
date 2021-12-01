package deployment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/cbroglie/mustache"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func TemplatesDir() string {
	return filepath.Join("data", "deployment_templates")
}

type Config struct {
	ID        string                 `json:"id" binding:"required"`
	Variables map[string]interface{} `json:"vars,omitempty"`
}

func (c *Config) toDeploymentCreateRequest() (*models.DeploymentCreateRequest, error) {
	contents, err := c.templateContents()
	if err != nil {
		return nil, err
	}

	vars, err := c.vars(contents)
	if err != nil {
		return nil, err
	}

	ctxt := map[string]interface{}{
		"vars": vars,
	}
	tplStr, err := mustache.Render(string(contents), ctxt)
	if err != nil {
		return nil, fmt.Errorf("unable to read deployment template for configuration [%s]: %w", c.ID, err)
	}

	var tpl struct {
		Template models.DeploymentCreateRequest `json:"template"`
	}

	if err := json.Unmarshal([]byte(tplStr), &tpl); err != nil {
		return nil, fmt.Errorf("unable to decode deployment template for configuration [%s]: %w", c.ID, err)
	}

	return &tpl.Template, nil
}

func (c *Config) vars(contents []byte) (map[string]interface{}, error) {
	var tpl struct {
		Vars map[string]struct {
			Type    string      `json:"type"`
			Default interface{} `json:"default"`
		} `json:"vars"`
	}

	if err := json.Unmarshal(contents, &tpl); err != nil {
		return nil, fmt.Errorf("unable to decode deployment template for configuration [%s]: %w", c.ID, err)
	}

	// Validate
	for name, _ := range c.Variables {
		_, exists := tpl.Vars[name]
		if !exists {
			return nil, fmt.Errorf("undefined variable [%s] in configuration [%s]", name, c.ID)
		}
	}

	// Override
	var vars map[string]interface{}
	for name, value := range tpl.Vars {
		vars[name] = value.Default
		if v, exists := c.Variables[name]; exists {
			vars[name] = v
		}
	}

	return vars, nil
}

func (c *Config) templateContents() ([]byte, error) {
	path := filepath.Join(TemplatesDir(), c.ID, "setup", "template.json")
	return ioutil.ReadFile(path)
}
