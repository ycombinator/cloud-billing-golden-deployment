package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/runners"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/dao"

	es "github.com/elastic/go-elasticsearch/v7"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"

	"github.com/gin-gonic/gin"
)

func registerScenarioRoutes(r *gin.Engine, scenarioRunner *runners.ScenarioRunner, stateConn *es.Client) {
	r.POST("/scenarios", postScenarios(scenarioRunner, stateConn))
	r.GET("/scenarios", getScenarios(stateConn))
	r.GET("/scenario/:id", getScenario(stateConn))
	r.DELETE("/scenario/:id")
}

func postScenarios(scenarioRunner *runners.ScenarioRunner, stateConn *es.Client) func(c *gin.Context) {
	scenarioDAO := dao.NewScenario(stateConn)
	return func(c *gin.Context) {
		var scenario models.Scenario
		if err := c.ShouldBindJSON(&scenario); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "could not parse scenario",
				"cause": err.Error(),
			})
			return
		}

		if err := scenario.GenerateID(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not generate ID for scenario",
				"cause": err.Error(),
			})
			return
		}

		if err := scenarioDAO.Save(&scenario); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not save scenario",
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

		if err := scenarioDAO.Save(&scenario); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"id":    scenario.ID,
				"error": "scenario started but could not be saved",
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
	scenarioDAO := dao.NewScenario(stateConn)
	return func(c *gin.Context) {
		scenarios, err := scenarioDAO.ListAll()
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

		scenarioDAO := dao.NewScenario(stateConn)
		scenario, err := scenarioDAO.Get(id)
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
