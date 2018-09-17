package server

import (
	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/utils"
)

func Init() {
	utils.ConnectToDb()

	if config.IsReleaseMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := setupRouter()
	r.Run(config.GetAddr())
}
