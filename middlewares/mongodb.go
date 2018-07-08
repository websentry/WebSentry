package middlewares

import (
    "log"

    "github.com/gin-gonic/gin"
    "gopkg.in/mgo.v2"
    "github.com/websentry/websentry/config"
)

var session *mgo.Session

var database string

func ConnectToDb() {
    c := config.GetMongodbConfig()
    database = c.Database

    s, err := mgo.Dial(c.Url)
    if err != nil {
        log.Fatal(err)
    }
    session = s
}

func MapDb(c *gin.Context) {
    s := session.Clone()
    defer s.Close()

    c.Set("mongo", s.DB(database))
    c.Next()
}
