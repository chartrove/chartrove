package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// generateToken is a helper to generate JWT tokens with flexible claims
func generateToken(user models.User, issuerSigningKey string, expiresAt time.Time, tokenID, tokenType string) (string, error) {
	if len(strings.TrimSpace(issuerSigningKey)) == 0 {
		return "", fmt.Errorf("issuer signing key cannot be empty")
	}
	claims := UserRegisteredClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "docker-fastenhealth",
			Subject:   user.Username,
			ID:        tokenID,
		},
		UserMetadata: UserMetadata{
			FullName: user.FullName,
			Email:    user.Email,
			Role:     user.Role,
		},
		TokenType: tokenType,
	}
	if user.ID != uuid.Nil {
		claims.NativeUserID = user.ID.String()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(issuerSigningKey))
}

// JwtGenerateFastenTokenFromUser generates a standard user token
func JwtGenerateFastenTokenFromUser(user models.User, issuerSigningKey string) (string, error) {
	return JwtGenerateFastenTokenWithLifetime(user, issuerSigningKey, time.Hour)
}

// JwtGenerateFastenTokenWithLifetime retains the standard user-token contract,
// changing only the absolute expiry of this newly issued session.
func JwtGenerateFastenTokenWithLifetime(user models.User, issuerSigningKey string, lifetime time.Duration) (string, error) {
	if lifetime < time.Second {
		return "", errors.New("session lifetime must be at least one second")
	}
	return generateToken(user, issuerSigningKey, time.Now().Add(lifetime), "", "")
}

func JwtValidateFastenToken(encryptionKey string, signedToken string) (*UserRegisteredClaims, error) {
	if len(strings.TrimSpace(encryptionKey)) == 0 {
		return nil, errors.New("issuer signing key cannot be empty")
	}
	// v5 skips issued-at checks by default; preserve v4 validation and require expiry.
	token, err := jwt.ParseWithClaims(
		signedToken,
		&UserRegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if jwt.SigningMethodHS256 != token.Method {
				return nil, fmt.Errorf("invalid signing algorithm: %s", token.Method)
			}
			return []byte(encryptionKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserRegisteredClaims)
	if !ok {
		err = errors.New("couldn't parse claims")
		return nil, err
	}
	return claims, nil
}

// JwtGenerateAccessToken generates an access token with custom expiration and metadata
func JwtGenerateAccessToken(user models.User, issuerSigningKey string, expiresAt time.Time, tokenID string) (string, error) {
	return generateToken(user, issuerSigningKey, expiresAt, tokenID, "access")
}
