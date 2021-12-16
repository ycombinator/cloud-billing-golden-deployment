package server

import (
	"fmt"
	"net/http"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/dao"

	es "github.com/elastic/go-elasticsearch/v7"

	"github.com/gin-gonic/gin"
)

func registerDeploymentConfigurationRoutes(r *gin.Engine, stateConn *es.Client) {
	r.PUT("/deployment_template/:id", putDeploymentConfiguration(stateConn))
	r.GET("/deployment_templates", getDeploymentConfigurations(stateConn))
	r.GET("/deployment_template/:id", getDeploymentConfiguration(stateConn))
	r.DELETE("/deployment_template/:id")
}

func putDeploymentConfiguration(stateConn *es.Client) func(c *gin.Context) {
	deploymentConfigDAO := dao.NewDeploymentConfiguration(stateConn)
	return func(c *gin.Context) {
		id := c.Param("id")
		var deploymentConfig models.DeploymentConfiguration

		if err := c.ShouldBindJSON(&deploymentConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not parse deployment configuration",
				"cause": err.Error(),
			})
			return
		}
		deploymentConfig.ID = id

		if err := deploymentConfigDAO.Save(&deploymentConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not save deployment configuration",
				"cause": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id": id,
			"resources": []string{
				fmt.Sprintf("/deployment_template/%s", id),
			},
		})
	}
}

func getDeploymentConfigurations(stateConn *es.Client) func(c *gin.Context) {
	deploymentConfigDAO := dao.NewDeploymentConfiguration(stateConn)
	return func(c *gin.Context) {
		deploymentConfigs, err := deploymentConfigDAO.ListAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read deployment configurations",
				"cause": err.Error(),
			})
			return
		}

		type item struct {
			ID        string   `json:"id"`
			Resources []string `json:"resources"`
		}

		var items []item
		for _, deploymentConfig := range deploymentConfigs {
			items = append(items, item{
				ID: deploymentConfig.ID,
				Resources: []string{
					fmt.Sprintf("/deployment_template/%s", deploymentConfig.ID),
				},
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"deployment_templates": items,
		})
	}
}

func getDeploymentConfiguration(stateConn *es.Client) func(c *gin.Context) {
	deploymentConfigDAO := dao.NewDeploymentConfiguration(stateConn)
	return func(c *gin.Context) {
		id := c.Param("id")

		deploymentConfig, err := deploymentConfigDAO.Get(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read deployment configuration",
				"cause": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, deploymentConfig)
	}
}
