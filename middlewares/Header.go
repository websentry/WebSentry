package middlewares

import (
	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/config"
)

var acao string

func init() {
	acao = config.GetAccessControlAllowOrigin()
}

func Header(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", acao)
	c.Next()
}
