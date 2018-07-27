package models

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
)

type Notification struct {
	Id bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	Name string `bson:"name" json:"name"`
	User bson.ObjectId `bson:"user" json:"-"`
	Type string `bson:"type" json:"type"`
	Setting map[string]interface{} `bson:"setting" json:"setting"`
}

func NotificationAddEmail(db *mgo.Database, user bson.ObjectId, email string, name string) (err error) {
	n := &Notification{}
	n.Name = name
	n.User = user
	n.Type = "email"
	n.Setting = gin.H{"email": email}

	err = db.C("Notifications").Insert(n)
	return
}

func NotificationCheckOwner(db *mgo.Database, id bson.ObjectId, user bson.ObjectId) bool {
	var result struct{ User bson.ObjectId `bson:"user"` }

	err := db.C("Notifications").Find(bson.M{"_id": id}).One(&result)
	if err==nil && result.User == user {
		return true
	}
	return false
}

func NotificationList(db *mgo.Database, user bson.ObjectId) (results []Notification, err error) {
	err = db.C("Notifications").Find(bson.M{"user": user}).All(&results)
	return
}