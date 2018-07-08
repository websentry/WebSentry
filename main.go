package main

import (
    "github.com/websentry/websentry/config"
    "github.com/websentry/websentry/server"
)

func main() {
    config.Init()
    server.Init()
}
