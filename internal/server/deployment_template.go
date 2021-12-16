package server

import (
	"fmt"
	"net/http"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/dao"

	es "github.com/elastic/go-elasticsearch/v7"

	"github.com/gin-gonic/gin"
)

func registerDeploymentTemplateRoutes(r *gin.Engine, stateConn *es.Client) {
	r.PUT("/deployment_template/:id", putDeploymentTemplate(stateConn))
	r.GET("/deployment_templates", getDeploymentTemplates(stateConn))
	r.GET("/deployment_template/:id", getDeploymentTemplate(stateConn))
	r.DELETE("/deployment_template/:id")
}

func putDeploymentTemplate(stateConn *es.Client) func(c *gin.Context) {
	deploymentTemplateDAO := dao.NewDeploymentTemplate(stateConn)
	return func(c *gin.Context) {
		id := c.Param("id")
		var deploymentTemplate models.DeploymentTemplate

		if err := c.ShouldBindJSON(&deploymentTemplate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not parse deployment template",
				"cause": err.Error(),
			})
			return
		}
		deploymentTemplate.ID = id

		if err := deploymentTemplateDAO.Save(&deploymentTemplate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not save deployment template",
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

func getDeploymentTemplates(stateConn *es.Client) func(c *gin.Context) {
	deploymentTemplateDAO := dao.NewDeploymentTemplate(stateConn)
	return func(c *gin.Context) {
		deploymentTemplates, err := deploymentTemplateDAO.ListAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read deployment templates",
				"cause": err.Error(),
			})
			return
		}

		type item struct {
			ID        string   `json:"id"`
			Resources []string `json:"resources"`
		}

		var items []item
		for _, deploymentTemplate := range deploymentTemplates {
			items = append(items, item{
				ID: deploymentTemplate.ID,
				Resources: []string{
					fmt.Sprintf("/deployment_template/%s", deploymentTemplate.ID),
				},
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"deployment_templates": items,
		})
	}
}

func getDeploymentTemplate(stateConn *es.Client) func(c *gin.Context) {
	deploymentTemplateDAO := dao.NewDeploymentTemplate(stateConn)
	return func(c *gin.Context) {
		id := c.Param("id")

		deploymentTemplate, err := deploymentTemplateDAO.Get(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read deployment template",
				"cause": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, deploymentTemplate)
	}
}
