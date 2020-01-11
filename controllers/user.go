package controllers

import (
	"math/rand"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/websentry/websentry/models"
	"github.com/websentry/websentry/utils"
)

const (
	minEmailLength = 3
	maxEmailLength = 254

	minPasswordLength = 8
	maxPasswordLength = 64

	verificationCodeLength      = 6
	maxVerificationCodeTryCount = 100
)

type fieldType int8

const (
	emailField fieldType = iota
	passwordField
	verficationCodeField
)

// UserInfo returns users' information, including email
func UserInfo(c *gin.Context) {
	result := models.User{}
	err := models.GetUserByID(c.MustGet("userId").(primitive.ObjectID), &result)
	if err != nil {
		panic(err)
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"email": result.Email,
	})
}

// UserLogin takes email and password and generate login token if succeed
func UserLogin(c *gin.Context) {
	gEmail := getFormattedEmail(c)
	gPassword := c.DefaultPostForm("password", "")

	if isFieldInvalid(gEmail, emailField) {
		JSONResponse(c, CodeWrongParam, "Email format is invalid", nil)
		return
	}

	if isFieldInvalid(gPassword, passwordField) {
		JSONResponse(c, CodeWrongParam, "Password format is invalid", nil)
		return
	}

	// check if the user exists
	userExist, err := models.CheckUserExistence(gEmail)
	if err != nil {
		panic(err)
	}
	if !userExist {
		JSONResponse(c, CodeNotExist, "sign up required", nil)
		return
	}

	// check password
	result := models.User{}
	err = models.GetUserByEmail(gEmail, &result)
	if err != nil {
		panic(err)
	}

	if !models.CheckPassword(gPassword, result.Password) {
		JSONResponse(c, CodeAuthError, "incorrect email/password", nil)
		return
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"token": utils.TokenGenerate(result.ID.Hex()),
	})
}

// UserGetSignUpVerification gets user email and password, generate Verification code and wait to be validated
func UserGetSignUpVerification(c *gin.Context) {
	gEmail := getFormattedEmail(c)

	// TODO: email check
	if isFieldInvalid(gEmail, emailField) {
		JSONResponse(c, CodeWrongParam, "Email format is invalid", nil)
		return
	}

	// check existence of the user
	userAlreadyExist, err := models.CheckUserExistence(gEmail)
	if err != nil {
		panic(err)
	}
	if userAlreadyExist {
		JSONResponse(c, CodeAlreadyExist, "", nil)
		return
	}

	var verificationCode string

	userVerificationExist, err := models.CheckUserVerificationExistence(gEmail)
	if err != nil {
		panic(err)
	}

	if userVerificationExist {
		// verfication code still valid
		JSONResponse(c, CodeOK, "", gin.H{
			"generated": false,
		})
	} else {
		verificationCode = generateVerificationCode()
		_, err = models.GetUserVerificationCollection().InsertOne(c, &models.UserVerification{
			Email:            gEmail,
			VerificationCode: verificationCode,
			RemainingCount:   maxVerificationCodeTryCount,
			CreatedAt:        time.Now(),
		})

		if err != nil {
			panic(err)
		}

		// we only send a verfication code once
		// until it is invalid due to exceeding limits of trying
		// or it expires
		utils.SendVerificationEmail(gEmail, verificationCode)

		JSONResponse(c, CodeOK, "", gin.H{
			"generated": true,
		})
	}
}

// UserCreateWithVerification checks verification code and create the user in the user database
func UserCreateWithVerification(c *gin.Context) {
	gEmail := getFormattedEmail(c)
	gPassword := c.DefaultPostForm("password", "")
	gVerificationCode := c.DefaultQuery("verification", "")

	if isFieldInvalid(gEmail, emailField) {
		JSONResponse(c, CodeWrongParam, "Email format is invalid", nil)
		return
	}

	if isFieldInvalid(gPassword, passwordField) {
		JSONResponse(c, CodeWrongParam, "Password format is invalid", nil)
		return
	}

	if isFieldInvalid(gVerificationCode, verficationCodeField) {
		JSONResponse(c, CodeWrongParam, "Verification format is invalid", nil)
		return
	}

	// check if it is already in the Users table
	userExist, err := models.CheckUserExistence(gEmail)
	if err != nil {
		panic(err)
	}

	if userExist {
		JSONResponse(c, CodeAlreadyExist, "", nil)
		return
	}

	// check if the user exist in UserVerifications table
	userVerificationExist, err := models.CheckUserVerificationExistence(gEmail)
	if err != nil {
		panic(err)
	}

	if !userVerificationExist {
		JSONResponse(c, CodeOK, "", nil)
		return
	}

	// check if the verification code is correct
	result := models.UserVerification{}
	err = models.GetUserVerificationByEmail(gEmail, &result)
	if err != nil {
		panic(err)
	}

	// exceed the trying limit
	if result.RemainingCount <= 0 {
		_, err = models.GetUserVerificationCollection().DeleteOne(c,
			bson.M{"email": gEmail},
		)
		if err != nil {
			panic(err)
		}

		JSONResponse(c, CodeAuthError, "exceed trying limit", gin.H{
			"expired": true,
		})
		return
	}

	// incorrect verification code
	if result.VerificationCode != gVerificationCode {
		// reduce remaining trying count
		_, err = models.GetUserVerificationCollection().UpdateOne(c,
			bson.M{"email": gEmail},
			bson.M{"$inc": bson.M{"remainingCount": -1}},
		)
		if err != nil {
			panic(err)
		}

		JSONResponse(c, CodeAuthError, "incorrect verification code", gin.H{
			"expired": false,
		})
		return
	}

	// insert to User table
	hash, err := models.HashPassword(gPassword)
	if err != nil {
		panic(err)
	}

	userID := primitive.NewObjectID()

	// insert doc containing "foreign key" first
	err = models.NotificationAddEmail(userID, gEmail, "--default--")

	if err != nil {
		panic(err)
	}

	_, err = models.GetUserCollection().InsertOne(c, &models.User{
		ID:          userID,
		Email:       gEmail,
		Password:    hash,
		TimeCreated: time.Now(),
	})

	if err != nil {
		panic(err)
	}

	JSONResponse(c, CodeOK, "", nil)
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

func isFieldInvalid(str string, field fieldType) bool {
	len := len(str)
	switch field {
	case emailField:
		return len < minEmailLength || len > maxEmailLength
	case passwordField:
		return len < minPasswordLength || len > maxPasswordLength
	case verficationCodeField:
		return len != verificationCodeLength
	default:
		return true
	}
}

func getFormattedEmail(c *gin.Context) string {
	return strings.ToLower(c.DefaultQuery("email", ""))
}
