package server

import (
	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/controllers"
	"github.com/websentry/websentry/middlewares"
	"github.com/websentry/websentry/models"
	"github.com/websentry/websentry/utils"
)

func Init(configFile string) error {

	err := config.Load(configFile)
	if err != nil {
		return err
	}

	db, err := connectToDB(config.GetConfig().Database)
	if err != nil {
		return err
	}

	err = models.Init(db)
	if err != nil {
		return err
	}

	controllers.Init()
	middlewares.Init()
	err = utils.Init()
	if err != nil {
		return err
	}

	if config.GetConfig().ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := setupRouter()

	r.ForwardedByClientIP = config.GetConfig().ForwardedByClientIP
	err = r.Run(config.GetConfig().Addr)
	return err
}
