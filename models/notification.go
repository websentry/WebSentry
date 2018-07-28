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

func GetNotification(db *mgo.Database, id bson.ObjectId) *Notification {
	c := db.C("Notifications")

	var result Notification
	err := c.Find(bson.M{"_id": id}).One(&result)
	if err!=nil {
		return nil
	}

	return &result
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

func NotificationAddServerChan(db *mgo.Database, name string, user bson.ObjectId, sckey string) (id bson.ObjectId, err error){
	n := &Notification{
		Id: bson.NewObjectId(),
		Name: name,
		User: user,
		Type: "serverchan",
		Setting: gin.H{"sckey": sckey},
	}

	err = db.C("Notifications").Insert(n)
	if err==nil {
		id = n.Id
	}
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



