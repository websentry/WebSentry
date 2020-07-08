package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/controllers"
)

var slaveKey string

func SlaveAuth(c *gin.Context) {
	if c.GetHeader("WS-Slave-Key") != slaveKey {

		controllers.JSONResponse(c, controllers.CodeAuthError, "", nil)

		c.Abort()
		return
	}
	c.Next()
}
