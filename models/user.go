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

func (t TX) CheckUserExistence(u string) (bool, error) {
	var count int64
	// TODO: not sure why the call to [Model] is required
	err := t.tx.Model(&User{}).Where(&User{Email: u}).Count(&count).Error
	return count == 1, err
}

// CheckEmailVerificationExistence finds out whether an user's verification code is already existed or not
// It takes a string represents the email
func (t TX) CheckEmailVerificationExistence(u string) (bool, error) {
	var count int64
	err := t.tx.Model(&EmailVerification{}).Where(&EmailVerification{Email: u}).Where("expired_at >= ?", time.Now()).Count(&count).Error
	return count == 1, err
}

// GetUserByEmail get the user's information
// It takes an email and a struct to store the result
func (t TX) GetUserByEmail(u string) (*User, error) {
	var result User
	err := t.tx.Where(&User{Email: u}).First(&result).Error

	if IsErrNoDocument(err) {
		return nil, nil
	}

	return &result, err
}

// GetEmailVerificationByEmail get the user's verification by email
// It takes an email and a struct to store the result
func (t TX) GetEmailVerificationByEmail(u string) (*EmailVerification, error) {
	var result EmailVerification
	err := t.tx.Where(&EmailVerification{Email: u}).Where("expired_at >= ?", time.Now()).First(&result).Error
	return &result, err
}

// GetUserByID get the user's information by his id,
// it takes a id, and a result structure
func (t TX) GetUserByID(id int64) (*User, error) {
	var result User
	err := t.tx.Where(&User{ID: id}).First(&result).Error
	if IsErrNoDocument(err) {
		return nil, nil
	}
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

// CreateCreateEmailVerification create new verfication code associated with an email address
func (t TX) CreateEmailVerification(u string) (string, error) {
	userExist, err := t.CheckUserExistence(u)
	if err != nil || userExist {
		return "", err
	}

	// Email column is unique, create will fail if an entry with the same email adress exists
	v := EmailVerification{
		Email:            u,
		VerificationCode: generateVerificationCode(),
		RemainingCount:   maxVerificationCodeTryCount,
		ExpiredAt:        time.Now().Add(time.Hour),
	}

	// TODO: create a new one or reuse the old one?
	err = t.tx.Create(&v).Error
	if err != nil {
		return "", err
	}

	return v.VerificationCode, err
}

func (t TX) DeleteEmailVerification(e *EmailVerification) error {
	return t.tx.Delete(&e).Error
}

func (t TX) UpdateEmailVerificationRemainingCount(e *EmailVerification) error {
	return t.tx.Select("remaining_count").Updates(e).Error
}

func (t TX) CreateUser(email string, pwdHash string) error {
	user := User{
		ID:       snowflakeNode.Generate().Int64(),
		Email:    email,
		Password: pwdHash,
	}
	err := t.tx.Create(&user).Error
	if err != nil {
		return err
	}
	return NotificationAddEmail(user.ID, email, "--default--")
}
