package models

import (

	"gopkg.in/mgo.v2/bson"
	"time"
	"gopkg.in/mgo.v2"
	"errors"
)

type SentryImage struct {
	Time time.Time `bson:"time"`
	File string `bson:"file"`
}

type Sentry struct {
	Id bson.ObjectId `bson:"_id,omitempty"`
	User bson.ObjectId `bson:"user"`
	CreateTime time.Time `bson:"createTime"`
	LastCheckTime time.Time `bson:"lastCheckTime"`
	NextCheckTime time.Time `bson:"nextCheckTime"`
	Interval int `bson:"interval"`
	CheckCount int `bson:"checkCount"`
	NotifyCount int `bson:"notifyCount"`
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

func getSentryVersionInterval(db *mgo.Database, id bson.ObjectId) (ver int, inter int, err error) {
	c := db.C("Sentries")

	var result struct{ Version int `bson:"version"`
						Interval int `bson:"interval"` }
	err = c.Find(bson.M{"_id": id}).One(&result)
	if err!=nil {
		return
	}
	ver = result.Version
	inter = result.Interval
	return
}

func UpdateSentryAfterCheck(db *mgo.Database, id bson.ObjectId, changed bool, newImage string, ver int) error {

	ver2, inter, err := getSentryVersionInterval(db, id)
	if err != nil {
		return err
	}

	if ver2!=ver {
		return errors.New("version changed")
	}

	c := db.C("Sentries")
	now := time.Now()

	up := bson.M{
			"$set": bson.M{"lastCheckTime": now,
							"nextCheckTime": now.Add(time.Minute * time.Duration(inter))},
			"$inc": bson.M{"checkCount": 1},
		}

	if changed {
		up["$inc"].(bson.M)["notifyCount"] = 1
		up["$set"].(bson.M)["image.time"] = now
		up["$set"].(bson.M)["image.file"] = newImage
	}

	err = c.Update(bson.M{"_id": id}, up)
	if err != nil {
		return err
	}

	return nil
}