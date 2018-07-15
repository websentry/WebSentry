package server

import (
    "github.com/websentry/websentry/config"
    "github.com/websentry/websentry/middlewares"
    "github.com/websentry/websentry/controllers"
)

func Init() {
    middlewares.ConnectToDb()

    // initialize verification email daemon
    controllers.VerificationEmailInit()

    r := setupRouter()
    r.Run(config.GetAddr())
}
