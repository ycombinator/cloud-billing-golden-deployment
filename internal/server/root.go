package server

import "github.com/gin-gonic/gin"

func registerRootRoute(r *gin.Engine) {
	r.GET("/", getRoot)
}

func getRoot(c *gin.Context) {
	c.JSON(200, gin.H{
		"resources": []string{
			"/deployment_configs",
			"/workloads",
			"/scenarios",
		},
	})
}
