package controllers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/models"
	"github.com/websentry/websentry/utils"
)

const (
	minEmailLength = 3
	maxEmailLength = 254

	minPasswordLength = 8
	maxPasswordLength = 64
)

type fieldType int8

const (
	emailField fieldType = iota
	passwordField
	verficationCodeField
)

// UserInfo returns users' information, including email
func UserInfo(c *gin.Context) {
	var userData *models.User

	err := models.Transaction(func(tx models.TX) (err error) {
		userData, err = tx.GetUserByID(c.MustGet("userId").(int64))
		return
	})
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	if userData == nil {
		JSONResponse(c, CodeNotExist, "", nil)
		return
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"email": userData.Email,
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

	var userID *int64
	err := models.Transaction(func(tx models.TX) (err error) {
		userID, err = tx.UserLogin(gEmail, gPassword)
		return
	})
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	if userID == nil {
		JSONResponse(c, CodeAuthError, "incorrect email/password", nil)
		return
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"token": utils.TokenGenerate(strconv.FormatInt(*userID, 16)),
	})
}

// UserGetSignUpVerification gets user email, generate Verification code and wait to be validated
func UserGetSignUpVerification(c *gin.Context) {
	gEmail := getFormattedEmail(c)

	// TODO: email check
	if isFieldInvalid(gEmail, emailField) {
		JSONResponse(c, CodeWrongParam, "Email format is invalid", nil)
		return
	}

	var vc string
	var userAlreadyExist, exceededLimit bool

	err := models.Transaction(func(tx models.TX) (err error) {
		userAlreadyExist, err = tx.CheckUserExistence(gEmail)
		if userAlreadyExist || err != nil {
			return
		}

		exceededLimit, err = tx.IsLastVerificationCodeGeneratedTimeExceeded(gEmail)
		if exceededLimit || err != nil {
			return
		}

		vc, err = tx.CreateEmailVerification(gEmail)
		return
	})
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	if userAlreadyExist {
		JSONResponse(c, CodeAlreadyExist, "", nil)
		return
	}

	if exceededLimit {
		JSONResponse(c, CodeExceededLimits, "", nil)
		return
	}

	// we only send a verfication code once
	// until it is invalid due to exceeding limits of trying
	// or it expires

	// TODO: handle the case where the email is failed to sent
	utils.SendVerificationEmail(gEmail, vc)

	// the user should not exist
	JSONResponse(c, CodeOK, "", nil)
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

	var correctVc, userAlreadyExist bool
	err := models.Transaction(func(tx models.TX) (err error) {
		correctVc, err = tx.CheckVerficationCode(gEmail, gVerificationCode)
		if !correctVc || err != nil {
			return
		}

		userAlreadyExist, err = tx.CheckUserExistence(gEmail)
		if userAlreadyExist || err != nil {
			return
		}

		err = tx.CreateUser(gEmail, gPassword)
		if err != nil {
			return
		}

		return
	})
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	if !correctVc {
		JSONResponse(c, CodeAuthError, "", nil)
		return
	}

	if userAlreadyExist {
		JSONResponse(c, CodeAlreadyExist, "", nil)
		return
	}

	JSONResponse(c, CodeOK, "", nil)
}

func isFieldInvalid(str string, field fieldType) bool {
	len := len(str)
	switch field {
	case emailField:
		return len < minEmailLength || len > maxEmailLength
	case passwordField:
		return len < minPasswordLength || len > maxPasswordLength
	case verficationCodeField:
		return len != models.VerificationCodeLength
	default:
		return true
	}
}

func getFormattedEmail(c *gin.Context) string {
	return strings.ToLower(c.DefaultQuery("email", ""))
}
