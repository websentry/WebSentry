package server

import (
	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/databases"
	"github.com/websentry/websentry/models"
	"log"
)

func Init() {

	db, err := databases.ConnectToMongoDB(config.GetMongodbConfig())
	if err != nil { log.Fatal(err) }

	err = models.Init(db)
	if err != nil { log.Fatal(err) }

	if config.IsReleaseMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := setupRouter()
	err = r.Run(config.GetAddr())
	if err != nil { log.Fatal(err) }
}
