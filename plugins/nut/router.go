package nut

import "github.com/gin-gonic/gin"

var _router = gin.Default()

// Router http router
func Router() *gin.Engine {
	return _router
}
