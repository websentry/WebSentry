package models

import (
	"errors"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"golang.org/x/crypto/bcrypt"
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
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Email       string        `bson:"email"`
	Password    string        `bson:"password"`
	TimeCreated time.Time     `bson:"createdAt"`

	// TODO: task id?
}

// CheckUserExistence finds out whether an user is already existed or not
// it takes an int represents which databases to search for
// (0: Users, 1: UserVerifications)
// and a string represents the email
func CheckUserExistence(dn int, u string) (bool, error) {

	c := GetUserCollection(dn)
	if c == nil {
		return false, errors.New("wrong parameter: databases does not exist")
	}

	count, err := c.Count(nil, bson.M{"email": u})

	if err != nil {
		return false, errors.New("failed to count")
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

// EnsureUserVerificationsIndex ensures the indecies of UserVerifications table are created
func EnsureUserVerificationsIndex() error {
	//TODO: waiting for reply from official google group

	// c := GetUserCollection(1)

	// set TTL
	//index := mgo.Index{
	//	Key:         []string{"createdAt"},
	//	ExpireAfter: expireTime,
	//}
	//index := mongo.IndexModel{
	//	Keys: bson.M{
	//		"createdAt": 1,
	//	},
	//	Options: mongo.NewIndexOptionsBuilder().ExpireAfterSeconds(expireTime).Build(),
	//}

	//_, err := c.Indexes().CreateOne(nil, index)

	return nil
}

// GetUserByEmail get the user's information in the desired table
// it takes a databases number:
// (0: Users, 1: UserVerifications)
// an email and a struct to store the result
func GetUserByEmail(dn int, u string, result interface{}) error {
	c := GetUserCollection(dn)
	return c.FindOne(nil, bson.M{"email": u}).Decode(result)
}

// GetUserById get the user's information by his id,
// it takes a id, and a result structure
func GetUserById(id primitive.ObjectID, result interface{}) error {
	c := GetUserCollection(0)
	return c.FindOne(nil, bson.M{"_id": id}).Decode(result)
}

// GetUserCollection gets the collection of the databases
// it takes a databases pointer and a databases number:
// (0: Users, 1: UserVerifications)
func GetUserCollection(dn int) *mongo.Collection {
	var c *mongo.Collection

	switch dn {
	case 0:
		c = mongoDB.Collection("Users")
	case 1:
		c = mongoDB.Collection("UserVerifications")
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
