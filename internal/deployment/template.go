package deployment

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cbroglie/mustache"

	"github.com/elastic/cloud-sdk-go/pkg/models"
)

func TemplatesDir() string {
	return filepath.Join("data", "deployment_templates")
}

type Template struct {
	ID        string                 `json:"id" binding:"required"`
	Variables map[string]interface{} `json:"vars,omitempty"`
}

func (t *Template) id() string {
	var id string
	id += "golden-" + t.ID

	keys := make([]string, len(t.Variables))
	for key, _ := range t.Variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var hashStrs []string
	for _, key := range keys {
		v := t.Variables[key]
		var val string
		switch v.(type) {
		case string:
			val = v.(string)
		case int:
			val = strconv.Itoa(v.(int))
		}

		str := key + "=" + val
		hashStrs = append(hashStrs, str)
	}

	hashStr := strings.Join(hashStrs, "|")
	m := md5.New()
	hashB := m.Sum([]byte(hashStr))

	id += "-" + hex.EncodeToString(hashB)

	return id
}

func (t *Template) toDeploymentCreateRequest() (*models.DeploymentCreateRequest, error) {
	contents, err := t.contents()
	if err != nil {
		return nil, err
	}

	vars, err := t.vars(contents)
	if err != nil {
		return nil, err
	}

	ctxt := map[string]interface{}{
		"vars": vars,
	}
	tplStr, err := mustache.Render(string(contents), ctxt)
	if err != nil {
		return nil, fmt.Errorf("unable to read deployment template for configuration [%s]: %w", t.ID, err)
	}

	var tpl struct {
		Template models.DeploymentCreateRequest `json:"template"`
	}

	if err := json.Unmarshal([]byte(tplStr), &tpl); err != nil {
		return nil, fmt.Errorf("unable to decode deployment template for configuration [%s]: %w", t.ID, err)
	}

	return &tpl.Template, nil
}

func (t *Template) vars(contents []byte) (map[string]interface{}, error) {
	var tpl struct {
		Vars map[string]struct {
			Type    string      `json:"type"`
			Default interface{} `json:"default"`
		} `json:"vars"`
	}

	if err := json.Unmarshal(contents, &tpl); err != nil {
		return nil, fmt.Errorf("unable to decode deployment template for configuration [%s]: %w", t.ID, err)
	}

	// Validate
	for name, _ := range t.Variables {
		_, exists := tpl.Vars[name]
		if !exists {
			return nil, fmt.Errorf("undefined variable [%s] in configuration [%s]", name, t.ID)
		}
	}

	// Override
	vars := make(map[string]interface{}, len(tpl.Vars))
	for name, value := range tpl.Vars {
		vars[name] = value.Default
		if v, exists := t.Variables[name]; exists {
			vars[name] = v
		}
	}

	return vars, nil
}

func (t *Template) contents() ([]byte, error) {
	path := filepath.Join(TemplatesDir(), t.ID, "setup", "template.json")
	return ioutil.ReadFile(path)
}
