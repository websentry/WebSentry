package databases

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/websentry/websentry/config"
)

func ConnectToMongoDB(dbConfig config.Mongodb) (*mongo.Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dbConfig.URL))
	if err != nil {
		return nil, err
	}

	err = client.Connect(nil)
	if err != nil {
		return nil, err
	}

	return client.Database(dbConfig.Database), nil
}
