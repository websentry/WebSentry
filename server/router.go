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
		v1.POST("/get_verification", controllers.UserGetSignUpVerification)
		v1.POST("/create_user", controllers.UserCreateWithVerification)
		v1.POST("/log_in", controllers.UserLogIn)

		// user
		userGroup := v1.Group("/user")
		userGroup.Use(middlewares.UserAuthRequired)
		{
			userGroup.GET("/info", func(c *gin.Context) {
				c.JSON(200, "Okay")
			})
		}

		// sentry
		sentryGroup := v1.Group("/sentry")
		{
			sentryGroup.POST("/request_full_screenshot", controllers.SentryRequestFullScreenshot)
			sentryGroup.POST("/wait_full_screenshot", controllers.SentryWaitFullScreenshot)
			sentryGroup.POST("/get_full_screenshot", controllers.SentryGetFullScreenshot)
			sentryGroup.POST("/create", controllers.SentryCreate)
		}

		// slave
		slaveGroup := v1.Group("/slave")
		slaveGroup.Use(middlewares.SlaveAuth)
		{
			slaveGroup.POST("/init", controllers.SlaveInit)
			slaveGroup.POST("/fetch_task", controllers.SlaveFetchTask)
			slaveGroup.POST("/submit_task", controllers.SlaveSubmitTask)
		}

		// common
		// commonGroup := v1.Group("/common")
		// {
		// }
	}

	return r
}
