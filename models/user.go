package models

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	expireTime     = time.Minute * 10
	encryptionCost = 14
)

// UserVerification : Entry in the Verification table
type UserVerification struct {
	Email            string    `bson:"email"`
	VerificationCode string    `bson:"verification"`
	CreatedAt        time.Time `bson:"createdAt"`
}

// User : Entry in the actual User table
type User struct {
	Id          bson.ObjectId `bson:"_id,omitempty"`
	Email       string        `bson:"email"`
	Password    string        `bson:"password"`
	TimeCreated time.Time     `bson:"createdAt"`

	// TODO: task id?
}

// CheckUserExistence finds out whether an user is already existed or not
// it takes a database pointer, an int represents which database to search for
// (0: Users, 1: UserVerifications)
// and a string represents the email
func CheckUserExistence(db *mgo.Database, dn int, u string) (bool, error) {

	c := GetUserCollection(db, dn)
	if c == nil {
		return false, errors.New("wrong parameter: database does not exist")
	}

	count, err := c.Find(bson.M{"email": u}).Count()

	if err != nil {
		return false, errors.New("failed to count")
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

// GetUserByEmail get the user's information in the desired table
// it takes a database pointer, a database number:
// (0: Users, 1: UserVerifications)
// an email and a struct to store the result
func GetUserByEmail(db *mgo.Database, dn int, u string, result interface{}) error {
	c := GetUserCollection(db, dn)
	return c.Find(bson.M{"email": u}).One(result)
}

// GetUserById get the user's information by his id,
// it takes a database pointer, a id, and a result structure
func GetUserById(db *mgo.Database, id bson.ObjectId, result interface{}) error {
	c := GetUserCollection(db, 0)
	return c.Find(bson.M{"_id": id}).One(result)
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

// HashPassword encrypts the password
func HashPassword(p string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(p), encryptionCost)
	return string(bytes), err
}

// CheckPassword check if the password matches
func CheckPassword(p string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(p))
	return err == nil
}
