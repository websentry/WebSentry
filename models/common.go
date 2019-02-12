package models

import "github.com/mongodb/mongo-go-driver/mongo"

var mongoDB *mongo.Database

func Init(db *mongo.Database) error {
	mongoDB = db
	return nil
}
