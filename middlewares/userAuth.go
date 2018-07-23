package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"gopkg.in/mgo.v2/bson"
	"github.com/websentry/websentry/utils"
)

func UserAuthRequired(c *gin.Context) {
	t := c.Query("token")

	u, err := utils.TokenValidate(t)

	if err != nil {
		switch err {
		case utils.ErrorParseToken:
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "Authorization error: Failed to parse the token",
			})
		case utils.ErrorParseClaim:
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "Authorization error: Failed to parse the claim",
			})
		case utils.ErrorTokenMalformed:
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "Authorization error: Token is malformed",
			})
		case utils.ErrorTokenExpired:
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "Authorization error: Token is expired",
			})
		case utils.ErrorTokenRequired:
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "Authorization error: Token is required",
			})
		}
		c.Abort()
	} else {
		if u != "" {
			// success
			bsonId := bson.ObjectIdHex(u)

			c.Set("userId", bsonId)
			c.Next()
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "Authorization error: id can not be empty",
			})
			c.Abort()
		}
	}
}
