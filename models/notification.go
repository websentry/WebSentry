package models

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID      primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name    string                 `bson:"name" json:"name"`
	User    primitive.ObjectID     `bson:"user" json:"-"`
	Type    string                 `bson:"type" json:"type"`
	Setting map[string]interface{} `bson:"setting" json:"setting"`
}

func GetNotification(id primitive.ObjectID) (*Notification, error) {
	c := mongoDB.Collection("Notifications")

	var result Notification
	err := c.FindOne(nil, bson.M{"_id": id}).Decode(&result)
	return &result, err
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

func NotificationAddServerChan(name string, user primitive.ObjectID, sckey string) (id primitive.ObjectID, err error) {
	n := &Notification{
		ID:      primitive.NewObjectID(),
		Name:    name,
		User:    user,
		Type:    "serverchan",
		Setting: gin.H{"sckey": sckey},
	}

	_, err = mongoDB.Collection("Notifications").InsertOne(nil, n)
	if err == nil {
		id = n.ID
	}
	return
}

func NotificationCheckOwner(id primitive.ObjectID, user primitive.ObjectID) bool {
	var result struct {
		User primitive.ObjectID `bson:"user"`
	}

	err := mongoDB.Collection("Notifications").FindOne(nil, bson.M{"_id": id}).Decode(&result)
	if err == nil && result.User == user {
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
