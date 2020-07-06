package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/controllers"
	"github.com/websentry/websentry/middlewares"
)

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = config.GetConfig().CROSAllowOrigins
	corsConfig.AddAllowHeaders("WS-User-Token")
	corsConfig.AddAllowHeaders("WS-Slave-Key")
	r.Use(cors.New(corsConfig))

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	v1 := r.Group("/v1")
	{

		// sensitive api
		sensitive := v1.Group("")
		sensitive.Use(middlewares.GetSensitiveLimiter())
		{
			sensitive.POST("/get_verification", controllers.UserGetSignUpVerification)
			sensitive.POST("/create_user", controllers.UserCreateWithVerification)
			sensitive.POST("/login", controllers.UserLogin)
		}

		// general api
		general := v1.Group("")
		general.Use(middlewares.UserAuthRequired)
		general.Use(middlewares.GetGeneralLimiter())
		{
			// user
			userGroup := general.Group("/user")
			{
				userGroup.POST("/info", controllers.UserInfo)
			}

			// sentry
			sentryGroup := general.Group("/sentry")
			{
				sentryGroup.POST("/wait_full_screenshot", controllers.SentryWaitFullScreenshot)
				sentryGroup.POST("/create", controllers.SentryCreate)
				sentryGroup.POST("/list", controllers.SentryList)
				sentryGroup.POST("/info", controllers.SentryInfo)
				sentryGroup.POST("/remove", controllers.SentryRemove)

				screenshot := sentryGroup.Group("")
				screenshot.Use(middlewares.GetScreenshotLimiter())
				{
					screenshot.POST("/request_full_screenshot", controllers.SentryRequestFullScreenshot)
				}
			}

			// notification
			notificationGroup := general.Group("/notification")
			{
				notificationGroup.POST("/list", controllers.NotificationList)
				notificationGroup.POST("/add_serverchan", controllers.NotificationAddServerChan)
			}

		}

		// slave
		slaveGroup := v1.Group("/slave")
		slaveGroup.Use(middlewares.SlaveAuth)
		slaveGroup.Use(middlewares.GetSlaveLimiter())
		{
			slaveGroup.POST("/init", controllers.SlaveInit)
			slaveGroup.POST("/fetch_task", controllers.SlaveFetchTask)
			slaveGroup.POST("/submit_task", controllers.SlaveSubmitTask)
		}

		// common
		commonGroup := v1.Group("/common")
		{
			commonGroup.GET("/get_history_image", controllers.GetHistoryImage)
			commonGroup.GET("/get_full_screenshot_image", controllers.GetFullScreenshotImage)
		}

	}

	return r
}
