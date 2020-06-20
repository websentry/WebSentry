package models

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Trigger struct contains the config of what will trigger a notification
type Trigger struct {
	SimilarityThreshold float64 `bson:"similarityThreshold"`
}

// type SentryMode int

// const (
// 	SentryModeImageBased SentryMode = 0
// )

// Sentry struct is the main one for describing a sentry

type ImageHistory struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	Images []SentryImage      `bson:"images" json:"images"`
}

func GetUncheckedSentry() (*Sentry, error) {
	c := mongoDB.Collection("Sentries")

	now := time.Now()

	// delay selected sentry 10 min
	update := bson.M{"$set": bson.M{"nextCheckTime": now.Add(time.Minute * 10)}}

	// execute on a sentry that is due
	var result Sentry
	filter := bson.M{"nextCheckTime": bson.M{"$lte": now}}
	opts := options.FindOneAndUpdate().SetSort(bson.M{"nextCheckTime": 1}).SetReturnDocument(options.Before).SetUpsert(false)

	err := c.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&result)
	if IsErrNoDocument(err) {
		return nil, nil
	}
	return &result, err
}

func GetUserSentries(user primitive.ObjectID) (results []Sentry, err error) {
	cur, err := mongoDB.Collection("Sentries").Find(context.TODO(), bson.M{"user": user})
	if err == nil {
		err = getAllFromCursor(cur, &results)
	}
	return
}

func GetSentry(id primitive.ObjectID) (*Sentry, error) {
	c := mongoDB.Collection("Sentries")

	var result Sentry
	err := c.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result)
	return &result, err
}

func CreateSentryAndImageHistory(s *Sentry) error {
	// insert doc containing "foreign key" first
	_, err := mongoDB.Collection("ImageHistories").InsertOne(context.TODO(), &ImageHistory{
		ID:     s.ID,
		Images: []SentryImage{},
	})
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = mongoDB.Collection("Sentries").InsertOne(context.TODO(), s)
	return errors.WithStack(err)
}

func DeleteSentry(id primitive.ObjectID) error {
	_, err := mongoDB.Collection("Sentries").DeleteOne(context.TODO(), bson.M{"_id": id})
	return errors.WithStack(err)
}

func DeleteImageHistory(id primitive.ObjectID) error {
	_, err := mongoDB.Collection("ImageHistories").DeleteOne(context.TODO(), bson.M{"_id": id})
	return errors.WithStack(err)
}

func GetImageHistory(id primitive.ObjectID) (*ImageHistory, error) {
	c := mongoDB.Collection("ImageHistories")

	var result ImageHistory
	err := c.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result)
	return &result, err
}

func GetSentryName(id primitive.ObjectID) (name string, err error) {
	c := mongoDB.Collection("Sentries")

	var result struct {
		Name string `bson:"name"`
	}
	err = c.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return
	}
	name = result.Name
	return
}

func GetSentryNotification(id primitive.ObjectID) (nid primitive.ObjectID, err error) {
	c := mongoDB.Collection("Sentries")

	var result struct {
		Notification primitive.ObjectID `bson:"notification"`
	}
	err = c.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return
	}
	nid = result.Notification
	return
}

func UpdateSentryAfterCheck(id primitive.ObjectID, changed bool, newImage string) error {
	c := mongoDB.Collection("Sentries")

	var result struct {
		Interval   int       `bson:"interval"`
		CreateTime time.Time `bson:"createTime"`
	}

	err := c.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return errors.WithStack(err)
	}

	now := time.Now()
	t := (int(now.Sub(result.CreateTime).Minutes()) / result.Interval) + 1
	nextTime := result.CreateTime.Add(time.Minute * time.Duration(t*result.Interval))

	up := bson.M{
		"$set": bson.M{"lastCheckTime": now,
			"nextCheckTime": nextTime},
		"$inc": bson.M{"checkCount": 1},
	}

	if changed {
		// add history
		c := mongoDB.Collection("ImageHistories")
		_, err = c.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
			"$push": bson.M{"images": &SentryImage{Time: now, File: newImage}},
		})

		if err != nil {
			return errors.WithStack(err)
		}
	}

	if changed {
		up["$inc"].(bson.M)["notifyCount"] = 1
		up["$set"].(bson.M)["image.time"] = now
		up["$set"].(bson.M)["image.file"] = newImage
	}

	c = mongoDB.Collection("Sentries")
	_, err = c.UpdateOne(context.TODO(), bson.M{"_id": id}, up)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
