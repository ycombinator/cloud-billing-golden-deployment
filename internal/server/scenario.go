package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ycombinator/cloud-billing-golden-deployment/internal/models"

	"github.com/gin-gonic/gin"
)

func scenariosDir() string {
	return filepath.Join("data", "scenarios")
}

func registerScenarioRoutes(r *gin.Engine) {
	r.POST("/scenarios", postScenarios)
	r.GET("/scenarios", getScenarios)
	r.GET("/scenario/:id", getScenario)
	r.DELETE("/scenario/:id")
}

func postScenarios(c *gin.Context) {
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

func getScenarios(c *gin.Context) {
	files, err := os.ReadDir(scenariosDir())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not read scenarios",
			"cause": err.Error(),
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
				fmt.Sprintf("/scenario/%s", dirname),
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"scenarios": items,
	})
}

func getScenario(c *gin.Context) {
	id := c.Param("id")

	path := filepath.Join(scenariosDir(), id, "scenario.json")
	c.File(path)
}
