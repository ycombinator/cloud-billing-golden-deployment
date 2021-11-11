package server

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const deploymentConfigsDir = "./deployment_configs"

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

	c.JSON(200, gin.H{
		"deployment_configs": dirnames,
	})
}

func getDeploymentConfig(c *gin.Context) {
	c.JSON(200, gin.H{
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
