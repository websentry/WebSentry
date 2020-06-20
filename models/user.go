package models

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	encryptionCost = 14
)

// CheckUserExistence finds out whether an user is already existed or not
// It takes a string represents the email
func CheckUserExistence(u string) (bool, error) {
	var count int64
	err := db.Where(&User{Email: u}).Count(&count).Error
	return count == 1, err
}

// CheckUserVerificationExistence finds out whether an user's verification code is already existed or not
// It takes a string represents the email
func CheckUserVerificationExistence(u string) (bool, error) {
	var count int64
	err := db.Where(&EmailVerification{Email: u}).Count(&count).Error
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

// GetUserVerificationByEmail get the user's verification by email
// It takes an email and a struct to store the result
func GetUserVerificationByEmail(u string) (*EmailVerification, error) {
	var result EmailVerification
	err := db.Where(&EmailVerification{Email: u}).First(&result).Error
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
