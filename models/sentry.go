package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrSentryNotRunning      = errors.New("Sentry is not in running state.")
	ErrInvalidNotificationID = errors.New("Invalid notification ID.")
	ErrZeroAffectedRows      = errors.New("Zero affected rows.")
)

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
	err := t.notificationCheckOwner(s.NotificationID, s.UserID)
	if err != nil {
		if IsErrNoDocument(err) {
			return 0, ErrInvalidNotificationID
		} else {
			return 0, err
		}
	}
	return s.ID, t.tx.Create(s).Error
}

func (t TX) UpdateSentry(sid int64, uid int64, s *Sentry) error {
	if s.NotificationID != 0 {
		err := t.notificationCheckOwner(s.NotificationID, uid)
		if err != nil {
			if IsErrNoDocument(err) {
				return ErrInvalidNotificationID
			} else {
				return err
			}
		}
	}

	err := t.sentryCheckOwner(sid, uid)
	if err != nil {
		return err
	}
	s.ID = sid

	// TODO: update the "NextCheckTime" when setting "RunningState" to "Running"
	return t.tx.Model(s).Updates(s).Error
}

func (t TX) DeleteSentry(id int64, uid int64) error {
	err := t.sentryCheckOwner(id, uid)
	if err != nil {
		return err
	}
	return t.tx.Delete(&Sentry{ID: id}).Error
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

func (t TX) sentryCheckOwner(id int64, userID int64) error {
	var count int64
	err := t.tx.Model(&Sentry{}).Where(&Sentry{ID: id, UserID: userID}).Count(&count).Error
	if err == nil {
		if count != 1 {
			return gorm.ErrRecordNotFound
		}
	}
	return err
}
