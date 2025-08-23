package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	user1 := uuid.New()
	tokenSecret := "super secret secret"

	tests := []struct {
		name        string
		userID      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
		expectError bool
	}{
		{
			name:        "valid JWT creation",
			userID:      user1,
			tokenSecret: tokenSecret,
			expiresIn:   time.Hour,
			expectError: false,
		},
		{
			name:        "valid JWT with short expiration",
			userID:      user1,
			tokenSecret: tokenSecret,
			expiresIn:   time.Microsecond,
			expectError: false,
		},
		{
			name:        "valid JWT with empty secret",
			userID:      user1,
			tokenSecret: "",
			expiresIn:   time.Hour * 24 * 7,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MakeJWT(tt.userID, tt.tokenSecret, tt.expiresIn)
			if tt.expectError {
				if err != nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "secret"

	// Create a valid token for testing
	validToken, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("failed to create valid token: %v", err)
	}

	// Create an expired token
	expiredToken, err := MakeJWT(userID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	// Create a token with different secret
	differentSecretToken, err := MakeJWT(userID, "different-secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to create token with different secret: %v", err)
	}

	tests := []struct {
		name         string
		tokenString  string
		tokenSecret  string
		expectedID   uuid.UUID
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			tokenSecret: secret,
			expectedID:  userID,
			expectError: false,
		},
		{
			name:        "expired token",
			tokenString: expiredToken,
			tokenSecret: secret,
			expectedID:  userID,
			expectError: true,
		},
		{
			name:        "wrong secret",
			tokenString: differentSecretToken,
			tokenSecret: secret,
			expectedID:  userID,
			expectError: true,
		},
		{
			name:        "malformed token",
			tokenString: "invalid.jwt.token",
			tokenSecret: secret,
			expectedID:  userID,
			expectError: true,
		},
		{
			name:        "empty token",
			tokenString: "",
			tokenSecret: secret,
			expectedID:  userID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		fmt.Println("running test")
		t.Run(tt.name, func(t *testing.T) {
			userID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if userID != tt.expectedID {
				t.Errorf("expected user ID %s, got %s", tt.expectedID, userID)
			}
		})
	}
}
