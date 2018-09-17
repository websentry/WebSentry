package utils

import (
	"github.com/websentry/websentry/config"
	"gopkg.in/mgo.v2"
	"log"
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

func GetDBSession() *mgo.Session {
	return session.Clone()
}

func SessionToDB(s *mgo.Session) *mgo.Database {
	return s.DB(database)
}
