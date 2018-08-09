package server

import (
	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/middlewares"
	"github.com/gin-gonic/gin"
)

func Init() {
	middlewares.ConnectToDb()

	if config.IsReleaseMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := setupRouter()
	r.Run(config.GetAddr())
}
