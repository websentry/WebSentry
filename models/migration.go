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

		var dbVersion SystemSetting
		err = db.FirstOrCreate(&dbVersion, SystemSetting{Key: "db_version"}).Error
		if err != nil {
			return
		}

		dbVersionInt, _ := strconv.Atoi(dbVersion.Value)

		if dbVersionInt == 0 {
			err = tx.AutoMigrate(&User{}, &EmailVerification{}, &NotificationMethod{}, &Sentry{}, &SentryImage{})
			if err != nil {
				return
			}
		} else {
			if dbVersionInt == 1 {
				err = db.Model(&Sentry{}).Where("notify_count = ?", -1).Update("notify_count", 0).Error
				if err != nil {
					return
				}
				err = tx.AutoMigrate(&User{}, &NotificationMethod{}, &Sentry{})
				if err != nil {
					return
				}
			}
		}
		dbVersion.Value = "2"
		return tx.Save(&dbVersion).Error
	})
}
