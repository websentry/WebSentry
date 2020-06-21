package server

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/controllers"
	"github.com/websentry/websentry/models"
)

func Init() {

	db, err := connectToDB(config.GetConfig().Database)
	if err != nil {
		log.Fatal(err)
	}

	err = models.Init(db)
	if err != nil {
		log.Fatal(err)
	}

	controllers.Init()

	if config.GetConfig().ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := setupRouter()

	r.ForwardedByClientIP = config.GetConfig().ForwardedByClientIP
	err = r.Run(config.GetConfig().Addr)
	if err != nil {
		log.Fatal(err)
	}
}
