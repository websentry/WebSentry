package models

import (

	"gopkg.in/mgo.v2/bson"
	"time"
	"gopkg.in/mgo.v2"
)

type SentryImage struct {
	Time time.Time `bson:"time"`
	File string `bson:"file"`
}

type Sentry struct {
	Id bson.ObjectId `bson:"_id,omitempty"`
	CreateTime time.Time `bson:"createTime"`
	LastCheckTime time.Time `bson:"lastCheckTime"`
	NextCheckTime time.Time `bson:"nextCheckTime"`
	Interval int `bson:"interval"`
	CheckCount int `bson:"checkCount"`
	Version int `bson:"version"`
	Image SentryImage `bson:"image"`
	Task map[string]interface{} `bson:"task"`
}

func GetUncheckedSentry(db *mgo.Database) *Sentry {
	c := db.C("Sentries")

	now := time.Now()

	// delay selected sentry 10 min
	change := mgo.Change{
		Update: bson.M{"$set": bson.M{"nextCheckTime": now.Add(time.Minute * 15)}},
		ReturnNew: false,
	}

	// execute on a sentry that is due
	var result Sentry
	_, err := c.Find(bson.M{"nextCheckTime": bson.M{"$lte": now,},}).Sort("-nextCheckTime").Apply(change, &result)
	if err!=nil {
		return nil
	}

	return &result
}

func GetSentry(db *mgo.Database, id bson.ObjectId) *Sentry {
	c := db.C("Sentries")

	var result Sentry
	err := c.Find(bson.M{"_id": id}).One(&result)
	if err!=nil {
		return nil
	}

	return &result
}