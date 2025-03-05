package common

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v4"
)

var (
    // ErrInvalidToken is returned when a token is invalid
    ErrInvalidToken = errors.New("invalid token")
    
    // ErrExpiredToken is returned when a token is expired
    ErrExpiredToken = errors.New("token expired")
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
    jwt.RegisteredClaims
    Username string `json:"username"`
    Role     string `json:"role"`
}

// GenerateJWT generates a new JWT token
func GenerateJWT(username, role, secret string, expirationTime time.Duration) (string, error) {
    claims := JWTClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
        Username: username,
        Role:     role,
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// ValidateJWT validates a JWT token
func ValidateJWT(tokenString, secret string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
    
    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrExpiredToken
        }
        return nil, err
    }
    
    if !token.Valid {
        return nil, ErrInvalidToken
    }
    
    claims, ok := token.Claims.(*JWTClaims)
    if !ok {
        return nil, ErrInvalidToken
    }
    
    return claims, nil
}

// RefreshJWT refreshes a JWT token
func RefreshJWT(tokenString, secret string, expirationTime time.Duration) (string, error) {
    claims, err := ValidateJWT(tokenString, secret)
    if err != nil && !errors.Is(err, ErrExpiredToken) {
        return "", err
    }
    
    return GenerateJWT(claims.Username, claims.Role, secret, expirationTime)
}
