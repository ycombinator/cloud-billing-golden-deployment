package server

import "github.com/gin-gonic/gin"

func registerWorkloadRoutes(r *gin.Engine) {
	r.GET("/workloads")
	r.GET("/workload/:id")
	r.GET("/workload/:id/payload")
	r.DELETE("/workload/:id")
}
