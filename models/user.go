package models

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	encryptionCost = 14

	VerificationCodeLength      = 6
	maxVerificationCodeTryCount = 100
)

// CheckUserExistence finds out whether an user is already existed or not
// It takes a string represents the email
func CheckUserExistence(u string) (bool, error) {
	var count int64
	err := db.Where(&User{Email: u}).Count(&count).Error
	return count == 1, err
}

// CheckEmailVerificationExistence finds out whether an user's verification code is already existed or not
// It takes a string represents the email
func CheckEmailVerificationExistence(u string) (bool, error) {
	var count int64
	err := db.Where(&EmailVerification{Email: u}).Where("expired_at >= ?", time.Now()).Count(&count).Error
	return count == 1, err
}

// GetUserByEmail get the user's information
// It takes an email and a struct to store the result
// TODO: what should it return if the record not exists?
func GetUserByEmail(u string) (*User, error) {
	var result User
	err := db.Where(&User{Email: u}).First(&result).Error
	return &result, err
}

// GetEmailVerificationByEmail get the user's verification by email
// It takes an email and a struct to store the result
func GetEmailVerificationByEmail(u string) (*EmailVerification, error) {
	var result EmailVerification
	err := db.Where(&EmailVerification{Email: u}).Where("expired_at >= ?", time.Now()).First(&result).Error
	return &result, err
}

// GetUserByID get the user's information by his id,
// it takes a id, and a result structure
func GetUserByID(id int64) (*User, error) {
	var result User
	err := db.Where(&User{ID: id}).First(&result).Error
	return &result, err
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

// generateVerificationCode outputs a random 6-digit code
func generateVerificationCode() string {
	numBytes := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rst := make([]byte, VerificationCodeLength)

	for i := range rst {
		rst[i] = numBytes[r.Intn(len(numBytes))]
	}

	return string(rst)
}

func CreateEmailVerification(u string) (string, error) {
	v := EmailVerification{
		Email:            u,
		VerificationCode: generateVerificationCode(),
		RemainingCount:   maxVerificationCodeTryCount,
		ExpiredAt:        time.Now().Add(time.Hour),
	}
	// TODO: create a new one or reuse the old one?
	err := db.Create(&v).Error
	return v.VerificationCode, err
}

func DeleteEmailVerification(e *EmailVerification) error {
	return db.Delete(&e).Error
}

func UpdateEmailVerificationRemainingCount(e *EmailVerification) error {
	return db.Select("remaining_count").Updates(e).Error
}

func CreateUser(email string, pwdHash string) error {
	user := User{
		ID:       snowflakeNode.Generate().Int64(),
		Email:    email,
		Password: pwdHash,
	}
	err := db.Create(&user).Error
	if err != nil {
		return err
	}
	return NotificationAddEmail(user.ID, email, "--default--")
}
