package deployment

import (
	"fmt"
	"net/http"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/config"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/auth"
)

type OutVars struct {
	ClusterID string
}

func EnsureDeployment(cfg *config.Config, template Template) (OutVars, error) {
	fmt.Printf("ensuring deployment for configuration [%s]...\n", template.ID)
	var out OutVars

	apiCfg := api.Config{
		Host:       cfg.API.Url,
		Client:     new(http.Client),
		AuthWriter: auth.APIKey(cfg.API.Key),
	}

	fmt.Println(apiCfg)

	ess, err := api.NewAPI(apiCfg)
	if err != nil {
		return out, fmt.Errorf("unable to connect to Elastic Cloud API at [%s]: %w", cfg.API.Url, err)
	}

	req, err := template.toDeploymentCreateRequest()
	if err != nil {
		return out, fmt.Errorf("unable to create deployment create request from configuration [%s]: %w", template.ID, err)
	}

	// TODO: make idempotent
	req.Name = template.id()
	resp, err := deploymentapi.Create(deploymentapi.CreateParams{
		API:     ess,
		Request: req,
	})
	if err != nil {
		fmt.Println(err)
		return out, fmt.Errorf("unable to ensure deployment for configuration [%s]: %w", err)
	}

	out.ClusterID = *resp.ID
	return out, nil
}
