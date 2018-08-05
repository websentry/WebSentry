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
	Name string `bson:"Name"`
	User bson.ObjectId `bson:"user"`
	Notification bson.ObjectId `bson:"notification"`
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
	Id bson.ObjectId `bson:"_id,omitempty"`
	Images []SentryImage `bson:"images"`
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

func GetUserSentries(db *mgo.Database, user bson.ObjectId) (results []Sentry, err error) {
	err = db.C("Sentries").Find(bson.M{"user": user}).All(&results)
	return
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

func GetSentryName(db *mgo.Database, id bson.ObjectId) (name string, err error) {
	c := db.C("Sentries")

	var result struct{ Name string `bson:"Name"` }
	err = c.Find(bson.M{"_id": id}).One(&result)
	if err!=nil {
		return
	}
	name = result.Name
	return
}

func GetSentryNotification(db *mgo.Database, id bson.ObjectId) (nid bson.ObjectId, err error) {
	c := db.C("Sentries")

	var result struct{ Notification bson.ObjectId `bson:"notification"` }
	err = c.Find(bson.M{"_id": id}).One(&result)
	if err!=nil {
		return
	}
	nid = result.Notification
	return
}

func getSentryInterval(db *mgo.Database, id bson.ObjectId) (inter int, err error) {
	c := db.C("Sentries")

	var result struct{ Interval int `bson:"interval"` }
	err = c.Find(bson.M{"_id": id}).One(&result)
	if err!=nil {
		return
	}
	inter = result.Interval
	return
}

func UpdateSentryAfterCheck(db *mgo.Database, id bson.ObjectId, changed bool, newImage string) error {

	inter, err := getSentryInterval(db, id)
	if err != nil {
		return err
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

	if changed {
		// add history
		c = db.C("ImageHistories")
		c.Update(bson.M{"_id": id}, bson.M{
			"$push": bson.M{"images": &SentryImage{Time:now, File:newImage}},
		})
	}

	return nil
}