package models

import (
	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
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

func NotificationCheckOwner(id primitive.ObjectID, user primitive.ObjectID) bool {
	var result struct{ User primitive.ObjectID `bson:"user"` }

	err := mongoDB.Collection("Notifications").FindOne(nil, bson.M{"_id": id}).Decode(&result)
	if err==nil && result.User == user {
		return true
	}
	return false
}

func NotificationList(user primitive.ObjectID) (results []Notification, err error) {
	cur, err := mongoDB.Collection("Notifications").Find(nil, bson.M{"user": user})
	if err == nil {
		err = getAllFromCursor(cur, &results)
	}
	return
}



