package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var ErrSentryNotRunning = errors.New("Sentry is not in running state.")

// If there isn't an unchecked sentry, it returns (nil, nil, nil)
func (t TX) GetUncheckedSentry() (*Sentry, *SentryImage, error) {
	var sResult Sentry
	now := time.Now()
	err := t.tx.Where("next_check_time <= ?", now).Where("running_state", RSRunning).Order("next_check_time").First(&sResult).Error
	if err != nil {
		if IsErrNoDocument(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	// delay selected sentry 10 min
	err = t.tx.Model(&sResult).Update("next_check_time", now.Add(time.Minute*10)).Error
	if err != nil {
		return nil, nil, err
	}
	if sResult.LatestImageID == nil {
		return &sResult, nil, nil
	}
	var iResult SentryImage
	err = t.tx.First(&iResult, *sResult.LatestImageID).Error
	return &sResult, &iResult, err
}

func (t TX) GetUserSentries(userID int64) (results []Sentry, err error) {
	err = t.tx.Where(&Sentry{UserID: userID}).Find(&results).Error
	return
}

func (t TX) GetSentry(id int64) (*Sentry, error) {
	var result Sentry
	err := t.tx.First(&result, id).Error
	return &result, err
}

func (t TX) CreateSentry(s *Sentry) (int64, error) {
	s.ID = snowflakeNode.Generate().Int64()
	return s.ID, t.tx.Create(s).Error
}

func (t TX) DeleteSentry(id int64, uid int64) error {
	res := t.tx.Delete(&Sentry{ID: id, UserID: uid})
	err := res.Error
	if err == nil && res.RowsAffected == 0 {
		err = gorm.ErrRecordNotFound
	}
	return err
}

func (t TX) GetImageHistory(id int64) (results []SentryImage, err error) {
	err = t.tx.Where(&SentryImage{SentryID: id}).Order("created_at DESC").Find(&results).Error
	return
}

func (t TX) GetSentryName(id int64) (string, error) {
	var result Sentry
	err := t.tx.Select("name").First(&result, id).Error
	return result.Name, err
}

func (t TX) GetSentryNotification(id int64) (int64, error) {
	var result Sentry
	err := t.tx.Select("notification_id").First(&result, id).Error
	return result.NotificationID, err
}

func (t TX) UpdateSentryAfterCheck(id int64, changed bool, newImage string) error {

	var result Sentry
	err := t.tx.Select("interval, created_at, notify_count, check_count, last_check_time, running_state").First(&result, id).Error
	if err != nil {
		return err
	}

	if result.RunningState != RSRunning {
		return ErrSentryNotRunning
	}

	var sentry Sentry

	now := time.Now()
	tc := (int(now.Sub(result.CreatedAt).Minutes()) / result.Interval) + 1
	firstTime := result.LastCheckTime == nil
	sentry.LastCheckTime = &now
	sentry.NextCheckTime = result.CreatedAt.Add(time.Minute * time.Duration(tc*result.Interval))
	sentry.CheckCount = result.CheckCount + 1

	if changed {
		// add image history
		sentryImage := SentryImage{
			SentryID: id,
			File:     newImage,
		}
		err = t.tx.Create(&sentryImage).Error
		if err != nil {
			return err
		}
		sentry.LatestImageID = &sentryImage.ID
		if !firstTime {
			sentry.NotifyCount = result.NotifyCount + 1
		}
	}

	return t.tx.Model(&Sentry{ID: id}).Updates(&sentry).Error
}
