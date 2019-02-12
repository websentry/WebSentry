package models

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Notification struct {
	Id primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name string `bson:"name" json:"name"`
	User primitive.ObjectID `bson:"user" json:"-"`
	Type string `bson:"type" json:"type"`
	Setting map[string]interface{} `bson:"setting" json:"setting"`
}

func GetNotification(id primitive.ObjectID) *Notification {
	c := mongoDB.Collection("Notifications")

	var result Notification
	err := c.FindOne(nil, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil
	}

	return &result
}

func NotificationAddEmail(user primitive.ObjectID, email string, name string) (err error) {
	n := &Notification{}
	n.Name = name
	n.User = user
	n.Type = "email"
	n.Setting = gin.H{"email": email}

	_, err = mongoDB.Collection("Notifications").InsertOne(nil, n)
	return
}

func NotificationAddServerChan(name string, user primitive.ObjectID, sckey string) (id primitive.ObjectID, err error){
	n := &Notification{
		Id: primitive.NewObjectID(),
		Name: name,
		User: user,
		Type: "serverchan",
		Setting: gin.H{"sckey": sckey},
	}

	_, err = mongoDB.Collection("Notifications").InsertOne(nil, n)
	if err == nil {
		id = n.Id
	}
	return
}

func NotificationCheckOwner(db *mgo.Database, id primitive.ObjectID, user primitive.ObjectID) bool {
	var result struct{ User primitive.ObjectID `bson:"user"` }

	err := db.C("Notifications").Find(bson.M{"_id": id}).One(&result)
	if err==nil && result.User == user {
		return true
	}
	return false
}

func NotificationList(db *mgo.Database, user primitive.ObjectID) (results []Notification, err error) {
	err = db.C("Notifications").Find(bson.M{"user": user}).All(&results)
	return
}



