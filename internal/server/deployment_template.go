package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	es "github.com/elastic/go-elasticsearch/v7"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/deployment"

	"github.com/gin-gonic/gin"
)

func registerDeploymentTemplateRoutes(r *gin.Engine, stateConn *es.Client) {
	r.GET("/deployment_templates", getDeploymentTemplates(stateConn))
	r.GET("/deployment_template/:id", getDeploymentTemplate(stateConn))
	r.DELETE("/deployment_template/:id")
}

func getDeploymentTemplates(stateConn *es.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		files, err := os.ReadDir(deployment.TemplatesDir())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read deployment configurations",
				"cause": err.Error(),
			})
			return
		}

		var dirnames []string
		for _, file := range files {
			if !file.IsDir() {
				continue
			}

			dirnames = append(dirnames, file.Name())
		}

		type item struct {
			ID        string   `json:"id"`
			Resources []string `json:"resources"`
		}

		var items []item
		for _, dirname := range dirnames {
			items = append(items, item{
				ID: dirname,
				Resources: []string{
					fmt.Sprintf("/deployment_template/%s", dirname),
				},
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"deployment_templates": items,
		})
	}
}

func getDeploymentTemplate(stateConn *es.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		path := filepath.Join(deployment.TemplatesDir(), id, "setup", "template.json")
		c.File(path)
	}
}
