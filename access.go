package s3

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

type AccessClaims struct {
	jwt.RegisteredClaims
	AccessKey string `json:"ak,omitempty"`
	Username  string `json:"un,omitempty"`
}

func CreateTokenWithClaims(secret string, claims jwt.Claims) (string, error) {
	if secret == "" {
		secret = defaultSecret
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.Wrap(err, "signed token string")
	}
	return t, nil
}

func CreateAccessToken(secret string, key string, expires time.Duration) (string, error) {
	claims := &AccessClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires)),
		},
		key,
		"",
	}
	return CreateTokenWithClaims(secret, claims)
}
