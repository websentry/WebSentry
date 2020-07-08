package middlewares

import "github.com/websentry/websentry/config"

func Init() {
	// slaveAuth
	slaveKey = config.GetConfig().SlaveKey
}
