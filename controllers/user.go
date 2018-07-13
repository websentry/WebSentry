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
	validationCodeLength = 6
)

// UserGetSignUpValidation gets user email and password, generate validation code and wait to be validated
func UserGetSignUpValidation(c *gin.Context) {
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
	collection := c.MustGet("mongo").(*mgo.Database).C("UserValidations")

	// set TTL
	index := mgo.Index{
		Key: []string{"createdAt"},

		ExpireAfter: time.Second * 30,
	}
	if err = collection.EnsureIndex(index); err != nil {
		panic(err)
	}

	var validationCode string

	count, err := collection.Find(bson.M{"username": gUsername}).Count()
	if err != nil {
		panic(err)
	}

	// TODO: test
	if count != 0 {
		// fetched validation code before
		result := models.UserValidation{}
		err = collection.Find(bson.M{"username": gUsername}).One(&result)
		if err != nil {
		panic(err)
		}

		validationCode = result.ValidationCode
		err = collection.Update(
			bson.M{"username": gUsername},
			bson.M{"$set": bson.M{"createdAt": time.Now()}},
		)
	} else {
		validationCode = generateValidationCode()
		err = collection.Insert(&models.UserValidation{
			Username:       gUsername,
			ValidationCode: validationCode,
			CreatedAt:      time.Now(),
		})
	}
	if err != nil {
		panic(err)
	}

	// sendValidationEmail(gUsername, validationCode)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg": "OK",
	})
}

// UserCreateWithValidation checks validation code and create the user in the user database
func UserCreateWithValidation(c *gin.Context) {
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

// generateValidationCode outputs a random 6-digit code
func generateValidationCode() string {
	numBytes := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rst := make([]byte, validationCodeLength)

	for i := range rst {
		rst[i] = numBytes[r.Intn(len(numBytes))]
	}

	return string(rst)
}
