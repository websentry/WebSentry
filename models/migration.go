package models

import (
	"strconv"

	"gorm.io/gorm"
)

func migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.AutoMigrate(&SystemSetting{})
		if err != nil {
			return
		}

		dbVersion := SystemSetting{Key: "db_version"}
		err = tx.Where(&dbVersion).First(&dbVersion).Error
		if err == gorm.ErrRecordNotFound {
			tx.Create(&dbVersion)
			err = nil
		}
		if err != nil {
			return
		}

		dbVersionInt, _ := strconv.Atoi(dbVersion.Value)

		if dbVersionInt == 0 {
			err = tx.AutoMigrate(&User{}, &EmailVerification{}, &NotificationMethod{}, &Sentry{}, &SentryImage{})
			if err != nil {
				return
			}
		}
		dbVersion.Value = "1"
		return tx.Save(&dbVersion).Error
	})
}
