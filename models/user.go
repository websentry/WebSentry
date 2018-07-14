package models

import (
	"time"
)

// UserVerification : Entry in the Verification table
type UserVerification struct {
	Username       string    `bson:"username"`
	VerificationCode string    `bson:"verification"`
	CreatedAt      time.Time `bson:"createdAt"`
}

// User : Entry in the actual User table
// bcrypt
type User struct {
	Username    string    `bson:"username"`
	Password    string    `bson:"password"`
	TimeCreated time.Time `bson:"createdAt"`

	// TODO: task id?
}
