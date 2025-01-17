package auth

import (
	"errors"
	"time"

	"sample-exchange/backend/config"
	"sample-exchange/backend/models"

	"github.com/golang-jwt/jwt/v5"
)

var (
	cfg *config.Config

	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

func Init(c *config.Config) {
	cfg = c
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func GenerateTokenPair(user *models.User) (accessToken, refreshToken string, err error) {
	// Generate access token
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTAccessDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	claims.Type = "refresh"
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(cfg.JWTRefreshDuration))

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err = token.SignedString([]byte(cfg.JWTSecret))

	return accessToken, refreshToken, err
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}
