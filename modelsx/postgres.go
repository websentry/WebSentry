package modelsx

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() {
	fmt.Println(time.Now())
	fmt.Println("----------")
	db, err := gorm.Open(postgres.Open("postgres://postgres:784596@localhost:5432/postgres?sslmode=disable"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.Transaction(migrate)

	if err != nil {
		panic(err)
	}
}

func migrate(tx *gorm.DB) (err error) {
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
}

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
	DeletedAt *time.Time
}

type EmailVerification struct {
	Email            string `gorm:"type:varchar(255);primary_key"`
	VerificationCode string `gorm:"type:char(6)"` // verificationCodeLength
	RemainingCount   int
	ExpiredAt        time.Time `gorm:"index"`
}

type NotificationMethod struct {
	ID        int64  `gorm:"primary_key;auto_increment:false"` // use snowflake for this ID
	Name      string `gorm:"type:varchar(100)"`
	UserID    int64  `gorm:"index"` // foreignkey: User.ID
	Type      string `gorm:"type:varchar(16)"`
	Setting   string // json
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
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
	LatestImageID  uint   // foreignkey: SentryImage.ID
	Task           string // json
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

type SentryImage struct {
	ID        uint      `gorm:"primary_key"`
	SentryID  int64     `gorm:"index:sentryid_createdat"` // foreignkey: Sentry.ID
	File      string    `gorm:"type:varchar(40)"`
	CreatedAt time.Time `gorm:"index:sentryid_createdat"`
}
