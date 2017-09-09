package nut

import "github.com/gin-gonic/gin"

var router = gin.Default()

// Router http router
func Router() *gin.Engine {
	return router
}
