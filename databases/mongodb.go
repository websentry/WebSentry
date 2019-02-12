package databases

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/websentry/websentry/config"
)

func ConnectToMongoDB(dbConfig config.Mongodb) (*mongo.Database, error) {
	client, err := mongo.Connect(nil, dbConfig.Url)
	if err != nil { return nil, err }

	// test connection
	err = client.Ping(nil, nil)
	if err != nil { return nil, err }

	return client.Database(dbConfig.Database), nil
}
