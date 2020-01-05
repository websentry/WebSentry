package models

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDB *mongo.Database

func Init(db *mongo.Database) error {
	mongoDB = db

	return ApplyMigrations()
}

// ApplyMigrations update the database schema
func ApplyMigrations() error {
	key := "general"
	var general struct {
		ID        string `bson:"_id,omitempty"`
		DbVersion int    `bson:"dbVersion"`
	}

	c := mongoDB.Collection("Admin")
	err := c.FindOne(nil, bson.M{"_id": key}).Decode(&general)
	if err != nil && !IsErrNoDocument(err) {
		return err
	}

	if general.DbVersion == 0 {
		println("Apply migration from dbVersion 0 to 1\n")
		general.ID = key
		// Add Index for user
		// User - UserVerifications - Index
		const expireTimeInSec = 60 * 10
		index := mongo.IndexModel{
			Keys: bson.M{
				"createdAt": 1,
			},
			Options: options.Index().SetExpireAfterSeconds(expireTimeInSec),
		}
		_, err = GetUserVerificationCollection().Indexes().CreateOne(context.Background(), index)
		if err != nil {
			return err
		}

		// update sentry schema
		_, err = mongoDB.Collection("Sentries").UpdateMany(nil, bson.M{}, bson.M{
			"$set": bson.M{
				"trigger": bson.M{
					"similarityThreshold": 0.9999,
				},
			},
		})
		if err != nil {
			return err
		}
		general.DbVersion = 1
	}

	_, err = c.ReplaceOne(nil, bson.M{"_id": key}, &general, options.Replace().SetUpsert(true))
	return err
}

// Helper for MongoDB

func IsErrNoDocument(err error) bool {
	return err == mongo.ErrNoDocuments
}

// modify from mgo "func (iter *Iter) All(result interface{}) error"
func getAllFromCursor(cur *mongo.Cursor, result interface{}) error {
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
		panic("result argument must be a slice address")
	}
	slicev := resultv.Elem()
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0
	for cur.Next(nil) {
		if slicev.Len() == i {
			elemp := reflect.New(elemt)
			err := cur.Decode(elemp.Interface())
			if err != nil {
				return err
			}

			slicev = reflect.Append(slicev, elemp.Elem())
			slicev = slicev.Slice(0, slicev.Cap())
		} else {
			err := cur.Decode(slicev.Index(i).Addr().Interface())
			if err != nil {
				return err
			}
		}
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))

	return cur.Close(nil)
}
