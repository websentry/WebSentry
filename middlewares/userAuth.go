package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/controllers"
	"github.com/websentry/websentry/utils"
	"gopkg.in/mgo.v2/bson"
)

func UserAuthRequired(c *gin.Context) {
	t := c.Query("token")

	u, err := utils.TokenValidate(t)

	if err != nil {
		detail := ""
		switch err {
		case utils.ErrorParseToken:
			detail = "Failed to parse the token"
		case utils.ErrorParseClaim:
			detail = "Failed to parse the claim"
		case utils.ErrorTokenMalformed:
			detail = "Token is malformed"
		case utils.ErrorTokenExpired:
			detail = "Token is expired"
		case utils.ErrorTokenRequired:
			detail = "Token is required"
		}
		controllers.JsonResponse(c, controllers.CodeAuthError, detail, nil)
		c.Abort()
	} else {
		if u != "" {
			// success
			bsonId := bson.ObjectIdHex(u)

			c.Set("userId", bsonId)
			c.Next()
		} else {
			controllers.JsonResponse(c, controllers.CodeAuthError, "Uid can not be empty", nil)
			c.Abort()
		}
	}
}
