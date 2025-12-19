package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateToken(t *testing.T) {
	secretKey := "test-secret-key"
	jwtService := NewJWTService(secretKey)

	userID := "test-user-123"
	token, err := jwtService.GenerateToken(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := jwtService.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestJWTService_ValidateToken_Success(t *testing.T) {
	secretKey := "test-secret-key"
	jwtService := NewJWTService(secretKey)

	userID := "test-user-456"
	token, err := jwtService.GenerateToken(userID)
	require.NoError(t, err)

	claims, err := jwtService.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	secretKey := "test-secret-key"
	jwtService := NewJWTService(secretKey)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "malformed token",
			token: "invalid.token.here",
		},
		{
			name:  "random string",
			token: "not-a-jwt-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtService.ValidateToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.Equal(t, ErrInvalidToken, err)
		})
	}
}

func TestJWTService_ValidateToken_WrongSecret(t *testing.T) {
	jwtService1 := NewJWTService("secret-key-1")
	jwtService2 := NewJWTService("secret-key-2")

	userID := "test-user-789"
	token, err := jwtService1.GenerateToken(userID)
	require.NoError(t, err)

	claims, err := jwtService2.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestJWTService_ValidateToken_ExpiredToken(t *testing.T) {
	secretKey := "test-secret-key"
	jwtService := NewJWTService(secretKey)

	claims := Claims{
		UserID: "test-user-expired",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // истек час назад
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	validatedClaims, err := jwtService.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
	assert.Equal(t, ErrExpiredToken, err)
}

func TestJWTService_ValidateToken_WrongAlgorithm(t *testing.T) {
	secretKey := "test-secret-key"
	jwtService := NewJWTService(secretKey)

	claims := Claims{
		UserID: "test-user-wrong-alg",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	validatedClaims, err := jwtService.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestJWTService_TokenLifetime(t *testing.T) {
	secretKey := "test-secret-key"
	jwtService := NewJWTService(secretKey)

	userID := "test-user-lifetime"
	token, err := jwtService.GenerateToken(userID)
	require.NoError(t, err)

	claims, err := jwtService.ValidateToken(token)
	require.NoError(t, err)

	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	timeDiff := actualExpiry.Sub(expectedExpiry)
	assert.True(t, timeDiff < time.Second && timeDiff > -time.Second,
		"Token expiry time should be approximately 24 hours from now")
}
