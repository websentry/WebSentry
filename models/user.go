package models

import (
	"time"
)

// UserValidation : Entry in the Validation table
type UserValidation struct {
	Username       string    `bson:"username"`
	ValidationCode string    `bson:"validationCode"`
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
