package deployment

import (
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	cloudModels "github.com/elastic/cloud-sdk-go/pkg/models"
)

type Credentials struct {
	CloudID  string `json:"cloud_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type OutVars struct {
	DeploymentCredentials Credentials
	ClusterIDs            []string
}

func CreateDeployment(api *api.API, name string, req *cloudModels.DeploymentCreateRequest) (OutVars, error) {
	var out OutVars

	req.Name = name
	resp, err := deploymentapi.Create(deploymentapi.CreateParams{
		API:     api,
		Request: req,
	})
	if err != nil {
		return out, fmt.Errorf("unable to ensure deployment for configuration [%s]: %w", err)
	}

	out.ClusterIDs = getClusterIDs(resp.Resources)
	out.DeploymentCredentials = *getDeploymentCredentials(resp.Resources)

	return out, nil
}

func DeleteDeployment(api *api.API, id string) error {
	if _, err := deploymentapi.Delete(deploymentapi.DeleteParams{
		API:          api,
		DeploymentID: id,
	}); err != nil {
		return fmt.Errorf("unable to delete deployment [%s]: %w", id, err)
	}

	return nil
}

func CheckIfDeploymentExists(api *api.API, name string) (bool, error) {
	resp, err := deploymentapi.List(deploymentapi.ListParams{
		API: api,
	})
	if err != nil {
		return false, fmt.Errorf("unable to list deployments: %w", err)
	}

	for _, deployment := range resp.Deployments {
		if deployment.Name != nil && *deployment.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func getClusterIDs(resources []*cloudModels.DeploymentResource) []string {
	clusterIDs := make([]string, 0)
	for _, resource := range resources {
		if resource.ID != nil {
			clusterIDs = append(clusterIDs, *resource.ID)
		}
	}

	return clusterIDs
}

func getDeploymentCredentials(resources []*cloudModels.DeploymentResource) *Credentials {
	for _, resource := range resources {
		if resource.Credentials != nil && resource.Credentials.Username != nil && resource.Credentials.Password != nil &&
			resource.CloudID != "" {
			cred := new(Credentials)

			cred.Username = *resource.Credentials.Username
			cred.Password = *resource.Credentials.Password
			cred.CloudID = resource.CloudID

			return cred
		}
	}
	return nil
}
