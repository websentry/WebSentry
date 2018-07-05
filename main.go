package main

import (
    "os"
    "io"

    "github.com/gin-gonic/gin"
)

func setupLogger() {
    gin.DisableConsoleColor()
    f, _ := os.Create("gin.log")

    //gin.DefaultWriter = io.MultiWriter(f)

    // logs to the file and console at the same time
    gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}

func setupRouter() *gin.Engine {
	r := gin.New()
    r.Use(gin.Logger())
    r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
        c.String(200, "pong")
	})

    return r
}

func main() {

    //setupLogger()

    r := setupRouter()

    // 0.0.0.0:8080
    r.Run(":8080")
}
