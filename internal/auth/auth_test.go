package auth

import (
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer some-tokens-be-token-here")

	t.Logf("Headers: %v", headers)

	authHeader := headers.Get("Authorization")
	t.Logf("Authorization header value: %s", authHeader)

	check, err := GetBearerToken(headers)
	if err != nil {
		t.Logf("Error checking token")
	}
	t.Logf("Passed through function value: %s", check)

}
