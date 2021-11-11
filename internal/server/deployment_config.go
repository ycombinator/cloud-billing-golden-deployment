package server

import (
	"os"

	"github.com/gin-gonic/gin"
)

func registerDeploymentConfigRoutes(r *gin.Engine) {
	r.GET("/deployment_configs", getDeploymentConfigs)
	r.GET("/deployment_config/:id")
	r.DELETE("/deployment_config/:id")
}

func getDeploymentConfigs(c *gin.Context) {
	files, err := os.ReadDir("./deployment_configs")
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
