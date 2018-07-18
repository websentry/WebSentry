package models

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	expireTime = time.Minute * 1
)

// UserVerification : Entry in the Verification table
type UserVerification struct {
	Username         string        `bson:"username"`
	VerificationCode string        `bson:"verification"`
	CreatedAt        time.Time     `bson:"createdAt"`
}

// User : Entry in the actual User table
// bcrypt
type User struct {
	Username    string        `bson:"username"`
	Password    string        `bson:"password"`
	TimeCreated time.Time     `bson:"createdAt"`

	// TODO: task id?
}

// CheckUserExistence finds out whether an user is already existed or not
// it takes a database pointer, an int represents which database to search for
// (0: Users, 1: UserVerifications)
// and a string represents the username
func CheckUserExistence(db *mgo.Database, dn int, u string) (bool, error) {

	c := GetUserCollection(db, dn)
	if c == nil {
		return false, errors.New("Wrong parameter: databse does not exist")
	}

	count, err := c.Find(bson.M{"username": u}).Count()

	if err != nil {
		return false, errors.New("Failed to count")
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

// EnsureUserVerificationsIndex ensures the indecies of UserVerifications table are created
func EnsureUserVerificationsIndex(db *mgo.Database) error {
	c := GetUserCollection(db, 1)

	// set TTL
	index := mgo.Index{
		Key:         []string{"createdAt"},
		ExpireAfter: expireTime,
	}

	return c.EnsureIndex(index)
}

// GetUser get the user's information in the desired table
// it takes a database pointer, a database number:
// (0: Users, 1: UserVerifications)
// an username and a struct to store the result
func GetUser(db *mgo.Database, dn int, u string, result interface{}) error {
	c := GetUserCollection(db, dn)
	return c.Find(bson.M{"username": u}).One(result)
}

// GetUserCollection gets the collection of the database
// it takes a database pointer and a database number:
// (0: Users, 1: UserVerifications)
func GetUserCollection(db *mgo.Database, dn int) *mgo.Collection {
	var c *mgo.Collection

	switch dn {
	case 0:
		c = db.C("Users")
	case 1:
		c = db.C("UserVerifications")
	default:
		c = nil
	}
	return c
}

func Encrypt(s string) string {
	return s
}