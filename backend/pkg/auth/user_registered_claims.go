package auth

import "github.com/golang-jwt/jwt/v5"

type UserRegisteredClaims struct {
	UserMetadata
	jwt.RegisteredClaims

	// Immutable native identity; sub remains the username for wire compatibility.
	NativeUserID string `json:"native_user_id,omitempty"`

	// Optional fields for access tokens
	TokenID   string `json:"token_id,omitempty"`
	TokenType string `json:"token_type,omitempty"`
}
