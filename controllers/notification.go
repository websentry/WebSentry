package controllers

import (
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/models"
	"gopkg.in/mgo.v2"
)

func notificationToggle(sentryId bson.ObjectId, old string, new string) {
	fmt.Println("Notification")
	fmt.Println(sentryId)
}


func NotificationList(c *gin.Context) {
	results, err := models.NotificationList(c.MustGet("mongo").(*mgo.Database), c.MustGet("userId").(bson.ObjectId))
	if err!=nil {
		panic(err)
	}

	c.JSON(200, gin.H{
		"code":   0,
		"msg":    "OK",
		"notifications": results,
	})
}