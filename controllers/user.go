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
	result, err := models.GetUserByID(c.MustGet("userId").(int64))
	if err != nil {
		InternalErrorResponse(c, err)
		return
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
		InternalErrorResponse(c, err)
		return
	}
	if !userExist {
		JSONResponse(c, CodeNotExist, "sign up required", nil)
		return
	}

	// check password
	result, err := models.GetUserByEmail(gEmail)
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	if !models.CheckPassword(gPassword, result.Password) {
		JSONResponse(c, CodeAuthError, "incorrect email/password", nil)
		return
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"token": utils.TokenGenerate(strconv.FormatInt(result.ID, 16)),
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
		InternalErrorResponse(c, err)
		return
	}
	if userAlreadyExist {
		JSONResponse(c, CodeAlreadyExist, "", nil)
		return
	}

	userVerificationExist, err := models.CheckEmailVerificationExistence(gEmail)
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	if userVerificationExist {
		// verfication code still valid
		JSONResponse(c, CodeOK, "", gin.H{
			"generated": false,
		})
	} else {
		verificationCode, err := models.CreateEmailVerification(gEmail)

		if err != nil {
			InternalErrorResponse(c, err)
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
		InternalErrorResponse(c, err)
		return
	}

	if userExist {
		JSONResponse(c, CodeAlreadyExist, "", nil)
		return
	}

	// check if the user exist in UserVerifications table
	userVerificationExist, err := models.CheckEmailVerificationExistence(gEmail)
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	if !userVerificationExist {
		JSONResponse(c, CodeOK, "", nil)
		return
	}

	// check if the verification code is correct
	result, err := models.GetEmailVerificationByEmail(gEmail)
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	// exceed the trying limit
	if result.RemainingCount <= 0 {
		err = models.DeleteEmailVerification(result)
		if err != nil {
			InternalErrorResponse(c, err)
			return
		}

		JSONResponse(c, CodeAuthError, "exceed trying limit", gin.H{
			"expired": true,
		})
		return
	}

	// incorrect verification code
	if result.VerificationCode != gVerificationCode {
		// reduce remaining trying count
		result.RemainingCount -= 1
		err = models.UpdateEmailVerificationRemainingCount(result)
		if err != nil {
			InternalErrorResponse(c, err)
			return
		}

		JSONResponse(c, CodeAuthError, "incorrect verification code", gin.H{
			"expired": false,
		})
		return
	}

	// insert to User table
	hash, err := models.HashPassword(gPassword)
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	err = models.CreateUser(gEmail, hash)

	if err != nil {
		InternalErrorResponse(c, err)
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
