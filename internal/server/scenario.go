package server

import (
	"fmt"
	"net/http"
	"time"

	es "github.com/elastic/go-elasticsearch/v7"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"

	"github.com/gin-gonic/gin"
)

func registerScenarioRoutes(r *gin.Engine, scenarioRunner *models.ScenarioRunner, stateConn *es.Client) {
	r.POST("/scenarios", postScenarios(scenarioRunner, stateConn))
	r.GET("/scenarios", getScenarios(stateConn))
	r.GET("/scenario/:id", getScenario(stateConn))
	r.DELETE("/scenario/:id")
}

func postScenarios(scenarioRunner *models.ScenarioRunner, stateConn *es.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		var scenario models.Scenario
		if err := c.ShouldBindJSON(&scenario); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not parse scenario",
				"cause": err.Error(),
			})
			return
		}

		if err := scenario.GenerateID(stateConn); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not generate ID for scenario",
				"cause": err.Error(),
			})
			return
		}

		if err := scenarioRunner.Start(&scenario); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not start scenario",
				"cause": err.Error(),
			})
			return
		}

		now := time.Now()
		scenario.StartedOn = &now
		scenario.StoppedOn = nil

		if err := scenario.Persist(stateConn); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"id":    scenario.ID,
				"error": "scenario started but could not be persisted",
				"cause": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id": scenario.ID,
			"resources": []string{
				fmt.Sprintf("/scenario/%s", scenario.ID),
			},
		})
	}
}

func getScenarios(stateConn *es.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		scenarios, err := models.LoadAllScenarios(stateConn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read scenarios",
				"cause": err.Error(),
			})
			return
		}

		type item struct {
			ID        string   `json:"id"`
			Resources []string `json:"resources"`
		}

		var items []item
		for _, scenario := range scenarios {
			items = append(items, item{
				ID: scenario.ID,
				Resources: []string{
					fmt.Sprintf("/scenario/%s", scenario.ID),
				},
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"scenarios": items,
		})
	}
}

func getScenario(stateConn *es.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		scenario, err := models.LoadScenario(id, stateConn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read scenario",
				"cause": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, scenario)
	}
}
