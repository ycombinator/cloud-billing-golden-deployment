package deployment

import (
	"fmt"
	"net/http"
	"os"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/auth"
)

type OutVars struct {
	ClusterID string
}

func EnsureDeployment(cfg Template) (OutVars, error) {
	fmt.Printf("ensuring deployment for configuration [%s]...\n", cfg.ID)
	var out OutVars

	apiKey := os.Getenv("EC_GOLDEN_API_KEY")
	if apiKey == "" {
		return out, fmt.Errorf("unable to obtain Elastic Cloud API key from environment variable [EC_GOLDEN_API_KEY]")
	}

	ess, err := api.NewAPI(api.Config{
		Client:     new(http.Client),
		AuthWriter: auth.APIKey(apiKey),
	})
	if err != nil {
		return out, fmt.Errorf("unable to connect to Elastic Cloud API: %w", err)
	}

	req, err := cfg.toDeploymentCreateRequest()
	if err != nil {
		return out, fmt.Errorf("unable to create deployment create request from configuration [%s]: %w", cfg.ID, err)
	}

	// TODO: make idempotent
	resp, err := deploymentapi.Create(deploymentapi.CreateParams{
		API:     ess,
		Request: req,
	})
	if err != nil {
		return out, fmt.Errorf("unable to ensure deployment for configuration [%s]: %w", err)
	}

	out.ClusterID = *resp.ID
	return out, nil
}
