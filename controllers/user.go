package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"net/http"
	"time"
	"github.com/websentry/WebSentry/utils"
)

const (
	verificationCodeLength = 6
)

func UserLogIn(c *gin.Context) {
	gUsername := c.Query("username")
	gPassword := c.Query("password")
	db := c.MustGet("mongo").(*mgo.Database)

	// check if the user exists
	userExist, err := models.CheckUserExistence(db, 0, gUsername)
	if err != nil {
		panic(err)
	}
	if !userExist {
		c.JSON(http.StatusOK, gin.H{
			"code": -3,
			"msg":  "Record does not exist: sign up required",
		})
		return
	}

	// check password
	result := models.User{}
	err = models.GetUserByUsername(db, 0, gUsername, &result)
	if err != nil {
		panic(err)
	}

	if !models.CheckPassword(gPassword, result.Password) {
		c.JSON(http.StatusOK, gin.H{
			"code": -2,
			"msg":  "Wrong parameter: incorrect username/password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "OK",
		"token": utils.TokenGenerate(result.Id.Hex()),
	})
}

// UserGetSignUpVerification gets user email and password, generate Verification code and wait to be validated
func UserGetSignUpVerification(c *gin.Context) {
	gUsername := c.Query("username")
	db := c.MustGet("mongo").(*mgo.Database)

	// check existence of the user
	userAlreadyExist, err := models.CheckUserExistence(db, 0, gUsername)
	if err != nil {
		panic(err)
	}
	if userAlreadyExist {
		c.JSON(http.StatusOK, gin.H{
			"code": -2,
			"msg":  "Wrong parameter: User already exists",
		})
		return
	}

	if err = models.EnsureUserVerificationsIndex(db); err != nil {
		panic(err)
	}

	var verificationCode string

	userVerificationExist, err := models.CheckUserExistence(db, 1, gUsername)
	if err != nil {
		panic(err)
	}

	if userVerificationExist {
		// fetched verification code before
		result := models.UserVerification{}
		err = models.GetUserByUsername(db, 1, gUsername, &result)
		if err != nil {
			panic(err)
		}

		verificationCode = result.VerificationCode
		err = models.GetUserCollection(db, 1).Update(
			bson.M{"username": gUsername},
			bson.M{"$set": bson.M{"createdAt": time.Now()}},
		)
	} else {
		verificationCode = generateVerificationCode()
		err = models.GetUserCollection(db, 1).Insert(&models.UserVerification{
			Username:         gUsername,
			VerificationCode: verificationCode,
			CreatedAt:        time.Now(),
		})
	}
	if err != nil {
		panic(err)
	}

	utils.SendVerificationEmail(gUsername, verificationCode)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "OK",
	})
}

// UserCreateWithVerification checks verification code and create the user in the user database
func UserCreateWithVerification(c *gin.Context) {
	gUsername := c.Query("username")
	gPassword := c.Query("password")
	gVerificationCode := c.Query("verification")

	db := c.MustGet("mongo").(*mgo.Database)

	// check if it is already in the Users table
	userExist, err := models.CheckUserExistence(db, 0, gUsername)
	if err != nil {
		panic(err)
	}

	if userExist {
		c.JSON(http.StatusOK, gin.H{
			"code": -2,
			"msg":  "Wrong parameter: user already exists",
		})
		return
	}

	// check if the user exist in UserVerifications table
	userVerificationExist, err := models.CheckUserExistence(db, 1, gUsername)
	if err != nil {
		panic(err)
	}

	if !userVerificationExist {
		c.JSON(http.StatusOK, gin.H{
			"code": -3,
			"msg":  "Record does not exist",
		})
		return
	}

	// check if the verification code is correct
	result := models.UserVerification{}
	err = models.GetUserByUsername(db, 1, gUsername, &result)
	if err != nil {
		panic(err)
	}
	if result.VerificationCode != gVerificationCode {
		c.JSON(http.StatusOK, gin.H{
			"code": -2,
			"msg":  "Wrong parameter: verification codes do not match",
		})
		return
	}

	// insert to User table
	hash, err := models.HashPassword(gPassword)
	if err != nil {
		panic(err)
	}

	err = models.GetUserCollection(db, 0).Insert(&models.User{
		Username:    gUsername,
		Password:    hash,
		TimeCreated: time.Now(),
	})

	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "OK",
	})
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
