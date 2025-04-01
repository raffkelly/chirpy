package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPassValidateJWT(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, "tokensecret", time.Hour)
	if err != nil {
		t.Fatalf("error creating token with MakeJWT")
	}
	_, err = ValidateJWT(token, "tokensecret")
	if err != nil {
		t.Fatalf("error validating a proper token")
	}
}

func TestFailValidateJWT(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, "tokensecretfail", time.Hour)
	if err != nil {
		t.Fatalf("error creating token with MakeJWT")
	}
	_, err = ValidateJWT(token, "tokensecret")
	if err == nil {
		t.Fatalf("improperly validated bad token")
	}
}

func TestGetBearerToken(t *testing.T) {
	header := http.Header{}
	header.Set("Authorization", "Bearer tokenstring")
	tokenString, err := GetBearerToken(header)
	if err != nil {
		t.Fatalf("failed getting token from header")
	}
	if tokenString != "tokenstring" {
		t.Fatalf("wrong token retrieved from header")
	}
}
