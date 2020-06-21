package models

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

func GetNotification(id int64) (*NotificationMethod, error) {
	var result NotificationMethod
	err := db.First(&result, id).Error
	if err != nil {
		return nil, err
	} else {
		return &result, err
	}
}

func NotificationAddEmail(userID int64, email string, name string) (err error) {
	n := &NotificationMethod{}
	n.ID = snowflakeNode.Generate().Int64()
	n.Name = name
	n.UserID = userID
	n.Type = "email"
	data, err := json.Marshal(gin.H{"email": email})
	if err != nil {
		return
	}
	n.Setting = string(data)
	return db.Create(n).Error
}

func NotificationAddServerChan(name string, userID int64, sckey string) (id int64, err error) {
	data, err := json.Marshal(gin.H{"sckey": sckey})
	if err != nil {
		return
	}
	n := &NotificationMethod{
		ID:      snowflakeNode.Generate().Int64(),
		Name:    name,
		UserID:  userID,
		Type:    "serverchan",
		Setting: string(data),
	}

	return n.ID, db.Create(n).Error
}

func NotificationCheckOwner(id int64, userID int64) (bool, error) {
	var count int64
	err := db.Model(&NotificationMethod{}).Where(&NotificationMethod{ID: id, UserID: userID}).Count(&count).Error
	return count == 1, err
}

func NotificationList(userID int64) (results []NotificationMethod, err error) {
	err = db.Where(&NotificationMethod{UserID: userID}).Find(&results).Error
	return
}
