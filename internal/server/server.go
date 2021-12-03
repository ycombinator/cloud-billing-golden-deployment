package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"
)

func Start(scenarioRunner *models.ScenarioRunner) error {
	r := gin.Default()

	// Routes
	registerRootRoute(r)
	registerDeploymentTemplateRoutes(r)
	//registerWorkloadRoutes(r)
	registerScenarioRoutes(r, scenarioRunner)

	return r.Run("localhost:8111")
}
