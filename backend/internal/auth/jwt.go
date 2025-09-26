package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

var jwtSecret = []byte("change_me")

// SetJWTSecret allows setting the JWT secret from environment
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// Claims represents the JWT claims structure
type Claims struct {
	SessionID string `json:"sid"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

// GenerateToken creates a new JWT token for a session
func GenerateToken(sessionID string) (string, error) {
	claims := Claims{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates a JWT token and returns the session ID
func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.SessionID, nil
	}

	return "", errors.New("invalid token")
}

// GetSessionIDFromRequest extracts and validates the JWT token from the request
func GetSessionIDFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	// Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return ValidateToken(parts[1])
}

// IsTokenExpired checks if a token is expired
func IsTokenExpired(tokenString string) bool {
	_, err := ValidateToken(tokenString)
	return err != nil
}
