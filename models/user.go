package models

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const (
	encryptionCost = 14
)

// UserVerification : Entry in the Verification table
type UserVerification struct {
	Email            string    `bson:"email"`
	VerificationCode string    `bson:"verification"`
	RemainingCount   int       `bson:"remainingCount"`
	CreatedAt        time.Time `bson:"createdAt"`
}

// User : Entry in the actual User table
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Email       string             `bson:"email"`
	Password    string             `bson:"password"`
	TimeCreated time.Time          `bson:"createdAt"`

	// TODO: task id?
}

// CheckUserExistence finds out whether an user is already existed or not
// It takes a string represents the email
func CheckUserExistence(u string) (bool, error) {
	c := GetUserCollection()
	return checkExistenceInCollection(c, u)
}

// CheckUserVerificationExistence finds out whether an user's verification code is already existed or not
// It takes a string represents the email
func CheckUserVerificationExistence(u string) (bool, error) {
	c := GetUserVerificationCollection()
	return checkExistenceInCollection(c, u)
}

// GetUserByEmail get the user's information
// It takes an email and a struct to store the result
func GetUserByEmail(u string, result interface{}) error {
	return GetUserCollection().FindOne(context.TODO(), bson.M{"email": u}).Decode(result)
}

// GetUserVerificationByEmail get the user's verification by email
// It takes an email and a struct to store the result
func GetUserVerificationByEmail(u string, result interface{}) error {
	return GetUserVerificationCollection().FindOne(context.TODO(), bson.M{"email": u}).Decode(result)
}

// GetUserByID get the user's information by his id,
// it takes a id, and a result structure
func GetUserByID(id primitive.ObjectID, result interface{}) error {
	c := GetUserCollection()
	return c.FindOne(context.TODO(), bson.M{"_id": id}).Decode(result)
}

// GetUserCollection returns the collection of the User
func GetUserCollection() *mongo.Collection {
	return mongoDB.Collection("Users")
}

// GetUserVerificationCollection returns the collection of UserVerification
func GetUserVerificationCollection() *mongo.Collection {
	return mongoDB.Collection("UserVerifications")
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

func checkExistenceInCollection(c *mongo.Collection, u string) (bool, error) {
	count, err := c.CountDocuments(context.TODO(), bson.M{"email": u})

	if err != nil {
		return false, errors.New("failed to count")
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}
