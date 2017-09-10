package nut

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var _router *gin.Engine
var routerOnce sync.Once

// Router http router
func Router() *gin.Engine {
	routerOnce.Do(func() {
		// gin.SetMode(gin.ReleaseMode)
		_router = gin.Default()
	})
	return _router
}
