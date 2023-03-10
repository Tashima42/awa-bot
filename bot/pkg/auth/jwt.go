package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"time"
)

type JWTHelper struct {
	secret []byte
}

type Token struct {
	UserID string `json:"userId"`
}

func NewJWTHelper(secret []byte) *JWTHelper {
	return &JWTHelper{
		secret: secret,
	}
}

func (j *JWTHelper) GenerateToken(token Token) (string, error) {
	jwtToken := jwt.New(jwt.SigningMethodHS256)
	claims := jwtToken.Claims.(jwt.MapClaims)

	claims["userID"] = token.UserID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := jwtToken.SignedString(j.secret)

	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWTHelper) VerifyToken(tokenString string) (*Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error parsing jwt")
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "token expired")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &Token{
			UserID: claims["userID"].(string),
		}, nil
	} else {
		return nil, errors.New("invalid token")
	}
}
