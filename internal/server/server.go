package server

import (
	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/runners"
)

func Start(scenarioRunner *runners.ScenarioRunner, stateConn *es.Client) error {
	r := gin.Default()

	// Routes
	registerRootRoute(r)
	registerDeploymentConfigurationRoutes(r, stateConn)
	registerScenarioRoutes(r, scenarioRunner, stateConn)

	return r.Run("localhost:8111")
}
