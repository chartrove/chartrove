package auth_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/auth"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/stretchr/testify/require"
)

// Sign the established wire format independently of either JWT dependency.
func legacyWireToken(payload string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body := header + "." + base64.RawURLEncoding.EncodeToString([]byte(payload))
	mac := hmac.New(sha256.New, []byte(syntheticSigningKey))
	_, _ = mac.Write([]byte(body))
	return body + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func TestJwtValidatesLegacyHS256Wire(t *testing.T) {
	issued := time.Now().Add(-time.Minute).Unix()
	expires := time.Now().Add(time.Hour).Unix()
	payload := fmt.Sprintf(`{"iss":"docker-fastenhealth","sub":"synthetic-legacy-user","iat":%d,"exp":%d,"full_name":"Synthetic Legacy User","picture":"https://example.invalid/synthetic.png","email":"synthetic@example.invalid","role":"user","jti":"synthetic-jti","token_id":"synthetic-optional-id","token_type":"access"}`, issued, expires)
	claims, err := auth.JwtValidateFastenToken(syntheticSigningKey, legacyWireToken(payload))
	require.NoError(t, err)
	require.Equal(t, "synthetic-legacy-user", claims.Subject)
	require.Equal(t, "docker-fastenhealth", claims.Issuer)
	require.Equal(t, "Synthetic Legacy User", claims.FullName)
	require.Equal(t, "https://example.invalid/synthetic.png", claims.Picture)
	require.Equal(t, "synthetic@example.invalid", claims.Email)
	require.Equal(t, "user", string(claims.Role))
	require.Equal(t, "synthetic-jti", claims.ID)
	require.Equal(t, "synthetic-optional-id", claims.TokenID)
	require.Equal(t, "access", claims.TokenType)
	require.Equal(t, issued, claims.IssuedAt.Unix())
	require.Equal(t, expires, claims.ExpiresAt.Unix())
}

func TestJwtGeneratedWireRetainsLegacySignatureAndClaims(t *testing.T) {
	user := models.User{Username: "synthetic-user", FullName: "Synthetic User", Email: "synthetic@example.invalid", Role: "user"}
	for _, kind := range []string{"user", "access"} {
		t.Run(kind, func(t *testing.T) {
			now := time.Now()
			accessExpiry := now.Add(24 * time.Hour).Truncate(time.Second)
			var signed string
			var err error
			if kind == "user" {
				signed, err = auth.JwtGenerateFastenTokenFromUser(user, syntheticSigningKey)
			} else {
				signed, err = auth.JwtGenerateAccessToken(user, syntheticSigningKey, accessExpiry, "synthetic-access-id")
			}
			require.NoError(t, err)
			parts := strings.Split(signed, ".")
			require.Len(t, parts, 3)
			signature, err := base64.RawURLEncoding.DecodeString(parts[2])
			require.NoError(t, err)
			mac := hmac.New(sha256.New, []byte(syntheticSigningKey))
			_, err = mac.Write([]byte(parts[0] + "." + parts[1]))
			require.NoError(t, err)
			require.True(t, hmac.Equal(mac.Sum(nil), signature), "legacy HS256 consumers must verify the generated signature")
			headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
			require.NoError(t, err)
			var header map[string]interface{}
			require.NoError(t, json.Unmarshal(headerBytes, &header))
			require.Equal(t, "HS256", header["alg"])
			require.Equal(t, "JWT", header["typ"])
			payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)
			var payload map[string]interface{}
			require.NoError(t, json.Unmarshal(payloadBytes, &payload))
			require.Equal(t, "docker-fastenhealth", payload["iss"])
			require.Equal(t, "synthetic-user", payload["sub"])
			require.Equal(t, "Synthetic User", payload["full_name"])
			require.Equal(t, "synthetic@example.invalid", payload["email"])
			require.Equal(t, "user", payload["role"])
			issued, ok := payload["iat"].(float64)
			require.True(t, ok)
			expires, ok := payload["exp"].(float64)
			require.True(t, ok)
			require.WithinDuration(t, now, time.Unix(int64(issued), 0), 2*time.Second)
			if kind == "user" {
				require.WithinDuration(t, now.Add(time.Hour), time.Unix(int64(expires), 0), 2*time.Second)
				require.NotContains(t, payload, "token_type")
				require.NotContains(t, payload, "jti")
			} else {
				require.Equal(t, float64(accessExpiry.Unix()), expires)
				require.Equal(t, "access", payload["token_type"])
				require.Equal(t, "synthetic-access-id", payload["jti"])
			}
		})
	}
}
