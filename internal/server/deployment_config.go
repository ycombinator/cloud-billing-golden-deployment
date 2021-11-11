package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const deploymentConfigsDir = "./data/deployment_configs"

func registerDeploymentConfigRoutes(r *gin.Engine) {
	r.GET("/deployment_configs", getDeploymentConfigs)
	r.GET("/deployment_config/:id", getDeploymentConfig)
	r.GET("/deployment_config/:id/payload", getDeploymentConfigPayload)
	r.DELETE("/deployment_config/:id")
}

func getDeploymentConfigs(c *gin.Context) {
	files, err := os.ReadDir(deploymentConfigsDir)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "could not read deployment configurations",
			"cause": err,
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
				fmt.Sprintf("/deployment_config/%s", dirname),
				fmt.Sprintf("/deployment_config/%s/payload", dirname),
			},
		})
	}

	c.JSON(200, gin.H{
		"deployment_configs": items,
	})
}

func getDeploymentConfig(c *gin.Context) {
	dc := c.Param("id")

	c.JSON(200, gin.H{
		"id": dc,
		"resources": []string{
			c.Request.RequestURI + "/payload",
		},
	})
}

func getDeploymentConfigPayload(c *gin.Context) {
	dc := c.Param("id")

	path := filepath.Join(deploymentConfigsDir, dc, "setup", "main.tf")
	c.File(path)
}
