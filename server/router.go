package server

import (
    "github.com/gin-gonic/gin"
    "github.com/websentry/websentry/middlewares"
    "github.com/websentry/websentry/controllers"
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




    return r
}
