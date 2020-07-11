package middlewares

import "github.com/websentry/websentry/config"

func Init() {
	// workerAuth
	workerKey = config.GetConfig().WorkerKey
}
