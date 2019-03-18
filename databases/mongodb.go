package databases

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/websentry/websentry/config"
)

func ConnectToMongoDB(dbConfig config.Mongodb) (*mongo.Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dbConfig.Url))
	if err != nil { return nil, err }

	// test connection
	err = client.Ping(nil, nil)
	if err != nil { return nil, err }

	return client.Database(dbConfig.Database), nil
}
