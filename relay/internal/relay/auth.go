package relay

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	SID string `json:"sid"`
	TID string `json:"tid"`
	jwt.RegisteredClaims `json:"exp,omitempty"`
}

func VerifyToken(secret []byte, tokenStr string) (Claims, error) {
	claims := &Claims{}

	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return Claims{}, err
	}

	if !tok.Valid {
		return Claims{}, fmt.Errorf("token is invalid")
	}

	if claims.RegisteredClaims.ExpiresAt != nil && claims.RegisteredClaims.ExpiresAt.Time.Before(time.Now()) {
		return Claims{}, fmt.Errorf("token expired")
	}

	return *claims, nil
}

func NewClaims(sid, tid string, exp time.Time) *Claims {
	return &Claims{
		SID: sid,
		TID: tid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
}
