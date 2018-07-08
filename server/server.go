package server

import (
    "github.com/websentry/websentry/config"
)

func Init() {
    r := setupRouter()
    r.Run(config.GetAddr())
}
