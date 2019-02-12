package databases

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/websentry/websentry/config"
)

var MongoDB *mongo.Database

func ConnectToMongoDB(dbConfig config.Mongodb) error {
	client, err := mongo.Connect(nil, dbConfig.Url)
	if err != nil { return err }

	// test connection
	err = client.Ping(nil, nil)
	if err != nil { return err }

	MongoDB = client.Database(dbConfig.Database)
	return nil
}
