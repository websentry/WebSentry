package models

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	encryptionCost                  = 14
	verificationCodeValidDuration   = 10 * time.Minute
	verficationCodeGenerateDuration = 1 * time.Minute

	VerificationCodeLength = 6
)

func (t TX) CheckUserExistence(u string) (bool, error) {
	var count int64
	// TODO: not sure why the call to [Model] is required
	err := t.tx.Model(&User{}).Where(&User{Email: u}).Count(&count).Error
	return count == 1, err
}

func (t TX) UserLogin(u, p string) (*int64, error) {
	var result User
	err := t.tx.Where(&User{Email: u}).First(&result).Error
	if err != nil {
		if IsErrNoDocument(err) {
			return nil, nil
		}
		return nil, err
	}

	if !checkPassword(p, result.Password) {
		return nil, nil
	}

	return &result.ID, err
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

func hashPassword(p string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(p), encryptionCost)
	return string(bytes), err
}

// CheckPassword check if the password matches
func checkPassword(p string, hash string) bool {
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

// IsLastVerificationCodeGeneratedTimeExceeded checks if the current is within the duration allowed
// starting from when the latest verification code is created
func (t TX) IsLastVerificationCodeGeneratedTimeExceeded(u string) (bool, error) {
	var result EmailVerification
	err := t.tx.Where(&EmailVerification{Email: u}).Order("expired_at desc").First(&result).Error
	if err != nil {
		if IsErrNoDocument(err) {
			return false, nil
		}
		return false, err
	}

	threasholdTime := result.ExpiredAt.Add(-verificationCodeValidDuration).Add(verficationCodeGenerateDuration)
	diff := time.Since(threasholdTime)

	if diff < 0 {
		return true, nil
	}
	return false, nil
}

// CreateCreateEmailVerification create new verfication code associated with an email address
func (t TX) CreateEmailVerification(u string) (string, error) {
	v := EmailVerification{
		Email:            u,
		VerificationCode: generateVerificationCode(),
		ExpiredAt:        time.Now().Add(verificationCodeValidDuration),
	}

	// there may exist multiple verfication codes that are valid to the user
	err := t.tx.Create(&v).Error
	return v.VerificationCode, err
}

// CheckVerficationCode checks if the given verification code is one of the non-expired verification existed in the db
func (t TX) CheckVerficationCode(u string, vc string) (bool, error) {
	var result EmailVerification
	err := t.tx.Where(&EmailVerification{Email: u, VerificationCode: vc}).Where("expired_at >= ?", time.Now()).First(&result).Error
	if err != nil {
		if IsErrNoDocument(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateUser add a new user and the default notification method
func (t TX) CreateUser(u string, pwdHash string) error {
	h, err := hashPassword(pwdHash)
	if err != nil {
		return err
	}

	user := User{
		ID:       snowflakeNode.Generate().Int64(),
		Email:    u,
		Password: h,
	}

	err = t.tx.Create(&user).Error
	if err != nil {
		return err
	}

	return NotificationAddEmail(user.ID, u, "--default--")
}

// TODO: Cleanup tables withe ExpireAt fields
