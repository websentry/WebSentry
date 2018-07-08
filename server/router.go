package server

import (
    "github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
    r := gin.New()
    r.Use(gin.Logger())
    r.Use(gin.Recovery())

    r.GET("/ping", func(c *gin.Context) {
        c.String(200, "pong")
    })

    return r
}
