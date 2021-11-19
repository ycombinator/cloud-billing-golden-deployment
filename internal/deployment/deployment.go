package deployment

import (
	"fmt"
	"path/filepath"
)

func TemplatesDir() string {
	return filepath.Join("data", "deployment_templates")
}

type Config struct {
	ID        string                 `json:"id" binding:"required"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type OutVars struct {
	ClusterID string
}

func EnsureDeployment(cfg Config) (OutVars, error) {
	fmt.Printf("ensuring deployment for configuration [%s]...\n", cfg.ID)
	var out OutVars

	path := filepath.Join(TemplatesDir(), cfg.ID, "setup")
	wd, err := NewWorkDir(path)
	if err != nil {
		return out, err
	}

	if err := wd.Init(); err != nil {
		return out, err
	}

	if err := wd.Apply(); err != nil {
		return out, err
	}

	// TODO: unharcode
	out.ClusterID = "129b342e90d443c6986a4fed59a09c0a"

	return out, nil
}
