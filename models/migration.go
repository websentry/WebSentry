package models

import (
	"strconv"

	"gorm.io/gorm"
)

func migrate(rawDB *gorm.DB) error {
	return rawDB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.AutoMigrate(&SystemSetting{})
		if err != nil {
			return
		}

		var dbVersion SystemSetting
		err = tx.FirstOrCreate(&dbVersion, SystemSetting{Key: "db_version"}).Error
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
				err = tx.Model(&Sentry{}).Where("notify_count = ?", -1).Update("notify_count", 0).Error
				if err != nil {
					return
				}
				err = tx.AutoMigrate(&User{}, &NotificationMethod{}, &Sentry{})
				if err != nil {
					return
				}
				dbVersionInt = 2
			}

			if dbVersionInt == 2 {
				err = tx.AutoMigrate(&EmailVerification{})
				if err != nil {
					return
				}
				err = tx.Table("email_verifications").Migrator().DropColumn(&EmailVerification{}, "remaining_count")
				if err != nil {
					return
				}
				dbVersionInt = 3
			}
			if dbVersionInt == 3 {
				err = tx.AutoMigrate(&User{})
				if err != nil {
					return
				}
				err = tx.Model(&User{}).Where("1 = 1").Updates(&User{
					Language: "en-US",
					TimeZone: "Asia/Shanghai",
				}).Error
				if err != nil {
					return
				}
				dbVersionInt = 4
			}
			if dbVersionInt == 4 {
				err = tx.AutoMigrate(&Sentry{})
				if err != nil {
					return
				}
				err = tx.Model(&Sentry{}).Where("1 = 1").Updates(&Sentry{
					RunningState: RSRunning,
				}).Error
				if err != nil {
					return
				}
				// dbVersionInt = 5
			}
		}
		dbVersion.Value = "5"

		return tx.Save(&dbVersion).Error
	})
}
