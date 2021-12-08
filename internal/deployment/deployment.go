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

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type OutVars struct {
	ClusterIDs            []string
	DeploymentCredentials Credentials
}

func EnsureDeployment(cfg *config.Config, template Template, suffix string) (OutVars, error) {
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

	deploymentName := fmt.Sprintf("golden-%s", suffix)

	if err := deleteExistingDeployment(ess, deploymentName); err != nil {
		return out, fmt.Errorf("unable to delete existing deployment: %w", err)
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
	out.DeploymentCredentials = *getDeploymentCredentials(resp.Resources)

	return out, nil
}

func deleteExistingDeployment(api *api.API, name string) error {
	id, err := getExistingDeployment(api, name)
	if err != nil {
		return err
	}

	if id == "" {
		return nil
	}

	if _, err = deploymentapi.Delete(deploymentapi.DeleteParams{
		API:          api,
		DeploymentID: id,
	}); err != nil {
		return fmt.Errorf("unable to delete deployment [%s]: %w", id, err)
	}

	return nil
}

func getExistingDeployment(api *api.API, name string) (string, error) {
	resp, err := deploymentapi.List(deploymentapi.ListParams{
		API: api,
	})
	if err != nil {
		return "", fmt.Errorf("unable to list deployments: %w", err)
	}

	for _, deployment := range resp.Deployments {
		if deployment.Name != nil && *deployment.Name == name && deployment.ID != nil {
			return *deployment.ID, nil
		}
	}

	return "", nil
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

func getDeploymentCredentials(resources []*models.DeploymentResource) *Credentials {
	for _, resource := range resources {
		if resource.Credentials != nil && resource.Credentials.Username != nil && resource.Credentials.Password != nil {
			cred := new(Credentials)
			cred.Username = *resource.Credentials.Username
			cred.Password = *resource.Credentials.Password

			return cred
		}
	}

	return nil
}
