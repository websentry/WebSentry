package models

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (t TX) GetNotification(id int64) (*NotificationMethod, error) {
	var result NotificationMethod
	err := t.tx.First(&result, id).Error
	if err != nil {
		return nil, err
	} else {
		return &result, err
	}
}

func (t TX) NotificationAddEmail(userID int64, email string, name string) (err error) {
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
	return t.tx.Create(n).Error
}

func (t TX) NotificationAddServerChan(name string, userID int64, sckey string) (id int64, err error) {
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

	return n.ID, t.tx.Create(n).Error
}

func (t TX) notificationCheckOwner(id int64, userID int64) error {
	var count int64
	err := t.tx.Model(&NotificationMethod{}).Where(&NotificationMethod{ID: id, UserID: userID}).Count(&count).Error
	if err == nil {
		if count != 1 {
			return gorm.ErrRecordNotFound
		}
	}
	return err
}

func (t TX) NotificationList(userID int64) (results []NotificationMethod, err error) {
	err = t.tx.Where(&NotificationMethod{UserID: userID}).Find(&results).Error
	return
}
