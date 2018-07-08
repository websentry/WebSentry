package server

import (
    "github.com/gin-gonic/gin"
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



    return r
}
