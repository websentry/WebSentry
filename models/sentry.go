package models

import (
	"time"

	"gorm.io/gorm"
)

func GetUncheckedSentry() (*Sentry, *SentryImage, error) {
	var sResult Sentry
	now := time.Now()
	err := db.Where("next_check_time <= ?", now).Order("next_check_time").First(&sResult).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	// delay selected sentry 10 min
	err = db.Model(&sResult).Update("next_check_time", now.Add(time.Minute*10)).Error
	if err != nil {
		return nil, nil, err
	}
	if sResult.LatestImageID == nil {
		return &sResult, nil, nil
	}
	var iResult SentryImage
	err = db.First(&iResult, *sResult.LatestImageID).Error
	return &sResult, &iResult, err
}

func GetUserSentries(userID int64) (results []Sentry, err error) {
	err = db.Where(&Sentry{UserID: userID}).Find(&results).Error
	return
}

func GetSentry(id int64) (*Sentry, error) {
	var result Sentry
	err := db.First(&result, id).Error
	return &result, err
}

func CreateSentry(s *Sentry) (int64, error) {
	s.ID = snowflakeNode.Generate().Int64()
	return s.ID, db.Create(s).Error
}

func DeleteSentry(id int64) error {
	return db.Delete(&Sentry{ID: id}).Error
}

func GetImageHistory(id int64) (results []SentryImage, err error) {
	err = db.Where(&SentryImage{SentryID: id}).Order("created_at DESC").Find(&results).Error
	return
}

func GetSentryName(id int64) (string, error) {
	var result Sentry
	err := db.Select("name").First(&result, id).Error
	return result.Name, err
}

func GetSentryNotification(id int64) (int64, error) {
	var result Sentry
	err := db.Select("notification_id").First(&result, id).Error
	return result.NotificationID, err
}

func UpdateSentryAfterCheck(id int64, changed bool, newImage string) error {

	var result Sentry
	err := db.Select("interval, create_at, notify_count, check_count").First(&result, id).Error
	if err != nil {
		return err
	}

	var sentry Sentry

	now := time.Now()
	t := (int(now.Sub(result.CreatedAt).Minutes()) / result.Interval) + 1
	sentry.LastCheckTime = &now
	sentry.NextCheckTime = result.CreatedAt.Add(time.Minute * time.Duration(t*result.Interval))
	sentry.CheckCount = result.CheckCount + 1

	if changed {
		// add image history
		sentryImage := SentryImage{
			SentryID: id,
			File:     newImage,
		}
		err = db.Create(&sentryImage).Error
		if err != nil {
			return err
		}
		sentry.LatestImageID = &sentryImage.ID
		sentry.NotifyCount = result.NotifyCount + 1
	}

	return db.Model(&Sentry{ID: id}).Updates(&sentry).Error
}
