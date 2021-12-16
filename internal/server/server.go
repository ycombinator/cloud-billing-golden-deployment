package server

import (
	"time"

	es "github.com/elastic/go-elasticsearch/v7"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/logging"
	"github.com/ycombinator/cloud-billing-golden-deployment/internal/runners"
)

func Start(scenarioRunner *runners.ScenarioRunner, stateConn *es.Client) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(ginzap.Ginzap(logging.Logger, time.RFC3339, true))

	// Routes
	registerRootRoute(r)
	registerDeploymentConfigurationRoutes(r, stateConn)
	registerScenarioRoutes(r, scenarioRunner, stateConn)

	return r.Run("localhost:8111")
}
