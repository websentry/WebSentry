package middlewares

import (
	"github.com/websentry/websentry/config"
	"github.com/gin-gonic/gin"
)

var slaveKey string

func init() {
	slaveKey = config.GetSlaveKey()
}

func SlaveAuth(c *gin.Context) {
	if c.Query("key")!=slaveKey {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "Authorization error",
		})
		c.Abort()
		return
	}
	c.Next()
}