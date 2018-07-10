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
		v1.POST("/sign_up", controllers.SignUp)
		v1.POST("/new_user", controllers.UserCreate)

		// user
		// userGroup := v1.Group("/user")
		// {
		// }

		// slave
		// slaveGroup := v1.Group("/slave")
		// {
		// }

		// common
		// commonGroup := v1.Group("/common")
		// {
		// }
	}

	r.POST("/v1/sentry/request_full_screenshot", controllers.SentryRequestFullScreenshot)
    r.POST("/v1/sentry/get_full_screenshot", controllers.SentryGetFullScreenshot)

	return r
}
