package relay

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyToken(t *testing.T) {
	secret := []byte("secret")
	sid := "sess_test"
	tid := "t_demo"
	exp := time.Now().Add(time.Hour)

	claims := NewClaims(sid, tid, exp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}

	// Valid
	verified, err := VerifyToken(secret, tokenStr)
	if err != nil {
		t.Errorf("Expected valid token, got error: %v", err)
	}
	if verified.SID != sid || verified.TID != tid {
		t.Errorf("Expected SID %s TID %s, got %s %s", sid, tid, verified.SID, verified.TID)
	}

	// Expired
	expTime := time.Now().Add(-time.Hour)
	oldClaims := NewClaims(sid, tid, expTime)
	oldToken := jwt.NewWithClaims(jwt.SigningMethodHS256, oldClaims)
	oldTokenStr, _ := oldToken.SignedString(secret)
	_, err = VerifyToken(secret, oldTokenStr)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestInvalidToken(t *testing.T) {
	secret := []byte("secret")
	_, err := VerifyToken(secret, "invalid")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}
