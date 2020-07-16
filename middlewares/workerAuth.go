package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/controllers"
)

var workerKey string

func WorkerAuth(c *gin.Context) {
	if c.GetHeader("WS-Worker-Key") != workerKey {

		controllers.JSONResponse(c, controllers.CodeAuthError, "", nil)

		c.Abort()
		return
	}
	c.Next()
}
