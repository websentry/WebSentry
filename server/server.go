package server

import (
	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/middlewares"
	"github.com/websentry/WebSentry/utils"
)

func Init() {
	middlewares.ConnectToDb()

	// initialize verification email daemon and init token's key
	utils.VerificationEmailInit()
	utils.TokenInit()

	r := setupRouter()
	r.Run(config.GetAddr())
}
