package main

import "github.com/websentry/websentry/server"

func main() {
	// TODO: reorganize init steps, avoid using golang's [init].
	server.Init()
}
