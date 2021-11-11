package server

import "github.com/gin-gonic/gin"

func Start() error {
	r := gin.Default()

	// Routes
	registerDeploymentConfigRoutes(r)
	registerWorkloadRoutes(r)
	registerScenarioRoutes(r)

	return r.Run("localhost:8111")
}
