package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/websentry/websentry/config"
)

const (
	expireDuration = time.Hour * 24 * 30
)

type claim struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

var secreteKey []byte = nil
var ErrorTokenMalformed error
var ErrorTokenExpired error
var ErrorParseToken error
var ErrorParseClaim error
var ErrorTokenRequired error

func init() {
	if secreteKey == nil {
		secreteKey = []byte(config.GetTokenSecretKey())
	}

	// initialize errors
	ErrorTokenMalformed = errors.New("not even a token")
	ErrorTokenExpired = errors.New("token already expired")
	ErrorParseToken = errors.New("failed to parse the token")
	ErrorParseClaim = errors.New("failed to parse the claim")
	ErrorTokenRequired = errors.New("token is required")
}

// TokenGenerate output the JWT token and it takes a user id and
// returns a string representing the token
func TokenGenerate(id string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(expireDuration).Unix(),
		"iss": "websentry",
	})

	rst, err := token.SignedString(secreteKey)
	if err != nil {
		panic(err)
	}

	return rst
}

// TokenValidate validates the token. It takes a token and returns
// a string representing the user Id and an error if it occurs. Possible errors are:
// ErrorParseClaim, ErrorParseToken, ErrorTokenExpired, ErrorTokenMalformed
func TokenValidate(t string) (string, error) {
	if t == "" {
		return "", ErrorTokenRequired
	}

	token, err := jwt.ParseWithClaims(t, &claim{}, func(token *jwt.Token) (interface{}, error) {
		return secreteKey, nil
	})

	if token == nil {
		return "", ErrorParseToken
	}

	if token.Valid {
		if c, ok := token.Claims.(*claim); ok {
			return c.ID, nil
		}
		return "", ErrorParseClaim
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return "", ErrorTokenMalformed
		} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
			return "", ErrorTokenExpired
		}
	}

	return "", ErrorParseToken
}
