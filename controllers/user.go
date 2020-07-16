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

	var userInfo *models.User

	err := models.Transaction(func(tx models.TX) (err error) {
		userInfo, err = tx.GetUserByEmail(gEmail)
		return
	})
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	if userInfo == nil || !models.CheckPassword(gPassword, userInfo.Password) {
		JSONResponse(c, CodeAuthError, "incorrect email/password", nil)
		return
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"token": utils.TokenGenerate(strconv.FormatInt(userInfo.ID, 16)),
	})
}

// UserSendVerification gets user email, generate Verification code and wait to be validated
func UserSendVerification(c *gin.Context) {
	gEmail := getFormattedEmail(c)

	// TODO: email check
	if isFieldInvalid(gEmail, emailField) {
		JSONResponse(c, CodeWrongParam, "Email format is invalid", nil)
		return
	}

	var userAlreadyExist, userVerificationExist bool
	var verificationCode *string

	err := models.Transaction(func(tx models.TX) (err error) {
		verificationCode, err = tx.CreateEmailVerification(gEmail)
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

	if userVerificationExist {
		// verfication code still valid
		JSONResponse(c, CodeOK, "", gin.H{
			"generated": false,
		})
		return
	}

	// we only send a verfication code once
	// until it is invalid due to exceeding limits of trying
	// or it expires

	// TODO: handle the case where the email is failed to sent
	utils.SendVerificationEmail(gEmail, verificationCode)

	JSONResponse(c, CodeOK, "", gin.H{
		"generated": true,
	})
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

	var emailVerifyInfo *models.EmailVerification

	err := models.Transaction(func(tx models.TX) (err error) {
		emailVerifyInfo, err = tx.GetEmailVerificationByEmail(gEmail)
		if err != nil {
			return
		}

		if emailVerifyInfo.VerificationCode != gVerificationCode {
			incorrectPwd = true
			err = tx.UpdateEmailVerificationRemainingCount(emailVerifyInfo)
			return
		}
		incorrectPwd = false

		// insert to User table
		hash, err := models.HashPassword(gPassword)
		if err != nil {
			return
		}

		err = tx.CreateUser(gEmail, hash)
		if err != nil {
			return
		}

		// delete used verification code
		err = tx.DeleteEmailVerification(emailVerifyInfo)
		return
	})
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	// hide details to front-end
	if userExist || !userVerificationExist {
		JSONResponse(c, CodeAuthError, "", nil)
		return
	}

	if incorrectPwd {
		JSONResponse(c, CodeAuthError, "", nil)
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
