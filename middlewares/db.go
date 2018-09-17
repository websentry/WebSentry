package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/utils"
)


func MapDb(c *gin.Context) {
	s := utils.GetDBSession()
	defer s.Close()

	c.Set("mongo", utils.SessionToDB(s))
	c.Next()
}
