package server

import "github.com/gin-gonic/gin"

func registerScenarioRoutes(r *gin.Engine) {
	r.GET("/scenarios")
	r.GET("/scenario/:id")
	r.DELETE("/scenario/:id")
}
