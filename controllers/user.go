package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"net/http"
	"time"
)

const (
	verificationCodeLength = 6
)

// UserGetSignUpVerification gets user email and password, generate Verification code and wait to be validated
func UserGetSignUpVerification(c *gin.Context) {
	gUsername := c.Query("username")

	// check existence of the user
	userAlreadyExist, err := checkUserExistence(gUsername, c)
	if err != nil {
		panic(err)
	}
	if userAlreadyExist {
		c.JSON(http.StatusOK, gin.H{
			"code": -2,
			"msg": "User already exists",
		})
		return
	}

	// connect to db
	collection := c.MustGet("mongo").(*mgo.Database).C("UserVerifications")

	// set TTL
	index := mgo.Index{
		Key: []string{"createdAt"},

		ExpireAfter: time.Second * 30,
	}
	if err = collection.EnsureIndex(index); err != nil {
		panic(err)
	}

	var verificationCode string

	count, err := collection.Find(bson.M{"username": gUsername}).Count()
	if err != nil {
		panic(err)
	}

	// TODO: test
	if count != 0 {
		// fetched verification code before
		result := models.UserVerification{}
		err = collection.Find(bson.M{"username": gUsername}).One(&result)
		if err != nil {
		panic(err)
		}

		verificationCode = result.VerificationCode
		err = collection.Update(
			bson.M{"username": gUsername},
			bson.M{"$set": bson.M{"createdAt": time.Now()}},
		)
	} else {
		verificationCode = generateVerificationCode()
		err = collection.Insert(&models.UserVerification{
			Username:       gUsername,
			VerificationCode: verificationCode,
			CreatedAt:      time.Now(),
		})
	}
	if err != nil {
		panic(err)
	}

	// send

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg": "OK",
	})
}

// UserCreateWithVerification checks verification code and create the user in the user database
func UserCreateWithVerification(c *gin.Context) {
	// TODO
}

// checkUserExistence finds out whether an user is already existed or not
func checkUserExistence(u string, c *gin.Context) (bool, error) {
	collection := c.MustGet("mongo").(*mgo.Database).C("Users")

	count, err := collection.Find(bson.M{"username": u}).Count()

	if err != nil {
		return false, errors.New("Failed to count")
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

// generateVerificationCode outputs a random 6-digit code
func generateVerificationCode() string {
	numBytes := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rst := make([]byte, verificationCodeLength)

	for i := range rst {
		rst[i] = numBytes[r.Intn(len(numBytes))]
	}

	return string(rst)
}
