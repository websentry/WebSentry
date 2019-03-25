package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/controllers"
)

var slaveKey string

func init() {
	slaveKey = config.GetSlaveKey()
}

func SlaveAuth(c *gin.Context) {
	if c.GetHeader("WS-Slave-Key") != slaveKey {

		controllers.JsonResponse(c, controllers.CodeAuthError, "", nil)

		c.Abort()
		return
	}
	c.Next()
}