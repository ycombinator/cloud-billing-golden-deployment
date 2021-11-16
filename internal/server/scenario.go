package server

import (
	"fmt"
	"net/http"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"

	"github.com/gin-gonic/gin"
)

func registerScenarioRoutes(r *gin.Engine) {
	r.POST("/scenarios", createScenario)
	r.GET("/scenarios")
	r.GET("/scenario/:id")
	r.DELETE("/scenario/:id")
}

func createScenario(c *gin.Context) {
	var scenario models.Scenario
	if err := c.ShouldBindJSON(&scenario); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "could not parse scenario",
			"cause": err.Error(),
		})
		return
	}

	if err := scenario.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid scenario",
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

	if err := scenario.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not start scenario",
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
