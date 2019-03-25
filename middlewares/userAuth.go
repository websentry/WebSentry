package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/websentry/websentry/controllers"
	"github.com/websentry/websentry/utils"
)

func UserAuthRequired(c *gin.Context) {
	t := c.GetHeader("WS-User-Token")

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
			bsonId, err := primitive.ObjectIDFromHex(u)
			if err != nil {
				controllers.JsonResponse(c, controllers.CodeAuthError, "Uid is invalid", nil)
				c.Abort()
			}

			c.Set("userId", bsonId)
			c.Next()
		} else {
			controllers.JsonResponse(c, controllers.CodeAuthError, "Uid can not be empty", nil)
			c.Abort()
		}
	}
}
