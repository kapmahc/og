package nut

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Wrap wrap handler
func Wrap(f func(*gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if e := f(c); e != nil {
			log.Error(e)
			c.String(http.StatusInternalServerError, e.Error())
			c.Abort()
		}
	}
}
