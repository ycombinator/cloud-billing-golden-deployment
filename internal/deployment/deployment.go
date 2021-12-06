package deployment

import (
	"fmt"
	"net/http"

	"github.com/elastic/cloud-sdk-go/pkg/models"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/config"

	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/auth"
)

type OutVars struct {
	ClusterIDs []string
}

func EnsureDeployment(cfg *config.Config, template Template) (OutVars, error) {
	fmt.Printf("ensuring deployment for configuration [%s]...\n", template.ID)
	var out OutVars

	apiCfg := api.Config{
		Host:       cfg.API.Url,
		Client:     new(http.Client),
		AuthWriter: auth.APIKey(cfg.API.Key),
	}

	ess, err := api.NewAPI(apiCfg)
	if err != nil {
		return out, fmt.Errorf("unable to connect to Elastic Cloud API at [%s]: %w", cfg.API.Url, err)
	}

	deploymentName := template.id()

	clusterIDs, err := checkIfDeploymentExists(ess, deploymentName)
	if err != nil {
		return out, fmt.Errorf("unable to check if deployment [%s] already exists: %w", deploymentName, err)
	}
	if clusterIDs != nil && len(clusterIDs) > 0 {
		out.ClusterIDs = clusterIDs
		return out, nil
	}

	req, err := template.toDeploymentCreateRequest()
	if err != nil {
		return out, fmt.Errorf("unable to create deployment create request from configuration [%s]: %w", template.ID, err)
	}

	req.Name = deploymentName
	resp, err := deploymentapi.Create(deploymentapi.CreateParams{
		API:     ess,
		Request: req,
	})
	if err != nil {
		return out, fmt.Errorf("unable to ensure deployment for configuration [%s]: %w", err)
	}

	out.ClusterIDs = getClusterIDs(resp.Resources)
	return out, nil
}

func checkIfDeploymentExists(api *api.API, name string) ([]string, error) {
	resp, err := deploymentapi.List(deploymentapi.ListParams{
		API: api,
	})
	if err != nil {
		return nil, fmt.Errorf("unable list deployments: %w", err)
	}

	for _, deployment := range resp.Deployments {
		if deployment.Name == nil {
			continue
		}

		if *deployment.Name == name {
			return getClusterIDs(deployment.Resources), nil
		}
	}

	return nil, nil
}

func getClusterIDs(resources []*models.DeploymentResource) []string {
	clusterIDs := make([]string, 0)
	for _, resource := range resources {
		if resource.ID != nil {
			clusterIDs = append(clusterIDs, *resource.ID)
		}
	}

	return clusterIDs
}
