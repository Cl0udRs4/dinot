package encryption

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid authentication token")
	// ErrTokenExpired is returned when a token has expired
	ErrTokenExpired = errors.New("authentication token has expired")
	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errors.New("invalid signature")
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Secret is the secret key used for HMAC signing
	Secret []byte
	// TokenExpiration is the duration after which tokens expire
	TokenExpiration time.Duration
	// EnableJWT enables JWT authentication
	EnableJWT bool
}

// DefaultAuthConfig returns the default authentication configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Secret:          []byte("default-secret-change-in-production"),
		TokenExpiration: 24 * time.Hour,
		EnableJWT:       true,
	}
}

// Authenticator handles authentication operations
type Authenticator struct {
	config AuthConfig
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(config AuthConfig) *Authenticator {
	return &Authenticator{
		config: config,
	}
}

// GenerateHMAC generates an HMAC for the given data
func (a *Authenticator) GenerateHMAC(data []byte) ([]byte, error) {
	if len(a.config.Secret) == 0 {
		return nil, errors.New("authentication secret not configured")
	}

	h := hmac.New(sha256.New, a.config.Secret)
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// VerifyHMAC verifies an HMAC for the given data
func (a *Authenticator) VerifyHMAC(data, signature []byte) error {
	expectedMAC, err := a.GenerateHMAC(data)
	if err != nil {
		return err
	}

	if !hmac.Equal(signature, expectedMAC) {
		return ErrInvalidSignature
	}

	return nil
}

// Claims represents the JWT claims
type Claims struct {
	ClientID string `json:"client_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT for the given client ID and role
func (a *Authenticator) GenerateJWT(clientID, role string) (string, error) {
	if !a.config.EnableJWT {
		return "", errors.New("JWT authentication not enabled")
	}

	expirationTime := time.Now().Add(a.config.TokenExpiration)
	claims := &Claims{
		ClientID: clientID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "dinot-server",
			Subject:   clientID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.config.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyJWT verifies a JWT and returns the claims
func (a *Authenticator) VerifyJWT(tokenString string) (*Claims, error) {
	if !a.config.EnableJWT {
		return nil, errors.New("JWT authentication not enabled")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.config.Secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrTokenExpired
	}

	return claims, nil
}

// GenerateBasicAuth generates a basic auth token
func (a *Authenticator) GenerateBasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// VerifyBasicAuth verifies a basic auth token
func (a *Authenticator) VerifyBasicAuth(token, expectedUsername, expectedPassword string) error {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return err
	}

	auth := string(decoded)
	username, password, found := splitBasicAuth(auth)
	if !found {
		return ErrInvalidToken
	}

	if username != expectedUsername || password != expectedPassword {
		return ErrInvalidToken
	}

	return nil
}

// splitBasicAuth splits a basic auth string into username and password
func splitBasicAuth(auth string) (username, password string, ok bool) {
	for i := 0; i < len(auth); i++ {
		if auth[i] == ':' {
			return auth[:i], auth[i+1:], true
		}
	}
	return "", "", false
}
