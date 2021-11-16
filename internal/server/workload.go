package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const workloadsDir = "./data/workloads"

func registerWorkloadRoutes(r *gin.Engine) {
	r.GET("/workloads", getWorkloads)
	r.GET("/workload/:id", getWorkload)
	r.GET("/workload/:id/payload", getWorkloadPayload)
	r.DELETE("/workload/:id")
}

func getWorkloads(c *gin.Context) {
	files, err := os.ReadDir(workloadsDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not read workloads",
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
				fmt.Sprintf("/workload/%s", dirname),
				fmt.Sprintf("/workload/%s/payload", dirname),
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"workloads": items,
	})
}

func getWorkload(c *gin.Context) {
	dc := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"id": dc,
		"resources": []string{
			c.Request.RequestURI + "/payload",
		},
	})
}

func getWorkloadPayload(c *gin.Context) {
	dc := c.Param("id")

	path := filepath.Join(workloadsDir, dc, "ops.log")
	c.File(path)
}
