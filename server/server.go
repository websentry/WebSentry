package server

import (
    "github.com/websentry/websentry/config"
    "github.com/websentry/websentry/middlewares"
)

func Init() {
    middlewares.ConnectToDb()

    r := setupRouter()
    r.Run(config.GetAddr())
}
