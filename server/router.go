package server

import (
	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/controllers"
	"github.com/websentry/websentry/middlewares"
)

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middlewares.MapDb)

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	v1 := r.Group("/v1")
	{
		v1.POST("/get_validation", controllers.UserGetSignUpValidation)
		v1.POST("/create_user", controllers.UserCreateWithValidation)

		// user
		// userGroup := v1.Group("/user")
		// {
		// }

		// sentry
		sentryGroup := v1.Group("/sentry")
		{
			sentryGroup.POST("/request_full_screenshot", controllers.SentryRequestFullScreenshot)
			sentryGroup.POST("/get_full_screenshot", controllers.SentryGetFullScreenshot)
		}

		// slave
		// slaveGroup := v1.Group("/slave")
		// {
		// }

		// common
		// commonGroup := v1.Group("/common")
		// {
		// }
	}

	return r
}
