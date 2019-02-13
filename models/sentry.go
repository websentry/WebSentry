package models

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type SentryImage struct {
	Time time.Time `bson:"time"`
	File string `bson:"file"`
}

type Sentry struct {
	Id primitive.ObjectID `bson:"_id,omitempty"`
	Name string `bson:"Name"`
	User primitive.ObjectID `bson:"user"`
	Notification primitive.ObjectID `bson:"notification"`
	CreateTime time.Time `bson:"createTime"`
	LastCheckTime time.Time `bson:"lastCheckTime"`
	NextCheckTime time.Time `bson:"nextCheckTime"`
	Interval int `bson:"interval"`
	CheckCount int `bson:"checkCount"`
	NotifyCount int `bson:"notifyCount"`
	Image SentryImage `bson:"image"`
	Task map[string]interface{} `bson:"task"`
}

type ImageHistory struct {
	Id primitive.ObjectID `bson:"_id,omitempty"`
	Images []SentryImage `bson:"images"`
}

func GetUncheckedSentry() (*Sentry, error) {
	c := mongoDB.Collection("Sentries")

	now := time.Now()

	// delay selected sentry 10 min
	update := bson.M{"$set": bson.M{"nextCheckTime": now.Add(time.Minute * 10)}}

	// execute on a sentry that is due
	var result Sentry
	filter := bson.M{"nextCheckTime": bson.M{"$lte": now,},}
	opts := options.FindOneAndUpdate().SetSort(bson.M{"nextCheckTime": 1}).SetReturnDocument(options.Before).SetUpsert(false)

	err := c.FindOneAndUpdate(nil, filter, update, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &result, err
}

func GetUserSentries(user primitive.ObjectID) (results []Sentry, err error) {
	cur, err := mongoDB.Collection("Sentries").Find(nil, bson.M{"user": user})
	if err == nil {
		err = getAllFromCursor(cur, &results)
	}
	return
}

//func GetSentry(id primitive.ObjectID) *Sentry {
//	c := mongoDB.Collection("Sentries")
//
//	var result Sentry
//	err := c.FindOne(nil, bson.M{"_id": id}).Decode(&result)
//	if err != nil {
//		return nil
//	}
//
//	return &result
//}

func AddSentry(s *Sentry) error {
	// insert doc containing "foreign key" first
	_, err := mongoDB.Collection("ImageHistories").InsertOne(nil, &ImageHistory{Id: s.Id})
	if err != nil { return err }
	_, err = mongoDB.Collection("Sentries").InsertOne(nil, s)
	return err
}

func GetSentryName(id primitive.ObjectID) (name string, err error) {
	c := mongoDB.Collection("Sentries")

	var result struct{ Name string `bson:"Name"` }
	err = c.FindOne(nil, bson.M{"_id": id}).Decode(&result)
	if err!=nil {
		return
	}
	name = result.Name
	return
}

func GetSentryNotification(id primitive.ObjectID) (nid primitive.ObjectID, err error) {
	c := mongoDB.Collection("Sentries")

	var result struct{ Notification primitive.ObjectID `bson:"notification"` }
	err = c.FindOne(nil, bson.M{"_id": id}).Decode(&result)
	if err!=nil {
		return
	}
	nid = result.Notification
	return
}

func getSentryInterval(id primitive.ObjectID) (inter int, err error) {
	c := mongoDB.Collection("Sentries")

	var result struct{ Interval int `bson:"interval"` }
	err = c.FindOne(nil, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return
	}
	inter = result.Interval
	return
}

func UpdateSentryAfterCheck(id primitive.ObjectID, changed bool, newImage string) error {

	inter, err := getSentryInterval(id)
	if err != nil {
		return err
	}


	c := mongoDB.Collection("Sentries")
	now := time.Now()

	up := bson.M{
			"$set": bson.M{"lastCheckTime": now,
							"nextCheckTime": now.Add(time.Minute * time.Duration(inter))},
			"$inc": bson.M{"checkCount": 1},
		}

	if changed {
		// add history
		c = mongoDB.Collection("ImageHistories")
		_, err = c.UpdateOne(nil, bson.M{"_id": id}, bson.M{
			"$push": bson.M{"images": &SentryImage{Time:now, File:newImage}},
		})

		if err != nil {
			return err
		}
	}


	if changed {
		up["$inc"].(bson.M)["notifyCount"] = 1
		up["$set"].(bson.M)["image.time"] = now
		up["$set"].(bson.M)["image.file"] = newImage
	}

	_, err = c.UpdateOne(nil, bson.M{"_id": id}, up)
	if err != nil {
		return err
	}

	return nil
}