package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/controllers"
)

func Init() {

	db, err := connectToDB(config.GetConfig().Database)
	if err != nil {
		log.Fatal(err)
	}

	// err = models.Init(db)
	fmt.Print(db)
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
