package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	secret := "my-secret"
	userID := uuid.New()
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("unexpected error creating token: %v", err)
	}

	validatedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}

	if validatedID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, validatedID)
	}
}
