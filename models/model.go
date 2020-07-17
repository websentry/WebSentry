package models

import (
	"time"

	"gorm.io/gorm"
)

// We use [string] to store [JSON] in db. Since for now we don't need to query [JSON] in SQL and using [string] allows
// us to support more database.
// We use snowflake instead of auto increment for public visible ID.
// Try not to use gorm's associations unless it's necessary because I believe it doesn't support laziness.

type SystemSetting struct {
	Key   string `gorm:"primary_key;type:varchar(64)"`
	Value string
}

type User struct {
	ID        int64  `gorm:"primary_key;auto_increment:false"` // use snowflake for this ID
	Email     string `gorm:"type:varchar(255);unique_index"`   // lower case
	Password  string `gorm:"type:char(60)"`                    // bcrypt
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type EmailVerification struct {
	ID               uint      `gorm:"primary_key"`
	Email            string    `gorm:"type:varchar(255);index:email_expiredat"`
	VerificationCode string    `gorm:"type:char(6)"` // verificationCodeLength
	ExpiredAt        time.Time `gorm:"index:email_expiredat"`
}

type NotificationMethod struct {
	ID        int64  `gorm:"primary_key;auto_increment:false"` // use snowflake for this ID
	Name      string `gorm:"type:varchar(100)"`
	UserID    int64  `gorm:"index"` // foreignkey: User.ID
	Type      string `gorm:"type:varchar(16)"`
	Setting   string // json
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Sentry struct {
	ID             int64  `gorm:"primary_key;auto_increment:false"` // use snowflake for this ID
	Name           string `gorm:"type:varchar(100)"`
	UserID         int64  `gorm:"index"` // foreignkey: User.ID
	NotificationID int64  // foreignkey: NotificationMethod.ID
	Trigger        string // json
	LastCheckTime  *time.Time
	NextCheckTime  time.Time `gorm:"index"`
	Interval       int
	CheckCount     int
	NotifyCount    int
	LatestImageID  *uint  // foreignkey: SentryImage.ID
	Task           string // json
	CreatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type SentryImage struct {
	ID        uint      `gorm:"primary_key"`
	SentryID  int64     `gorm:"index:sentryid_createdat"` // foreignkey: Sentry.ID
	File      string    `gorm:"type:varchar(40)"`
	CreatedAt time.Time `gorm:"index:sentryid_createdat"`
}

// Stored as json string
type Trigger struct {
	SimilarityThreshold float64 `json:"similarityThreshold"`
}
