package token

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateAccessToken(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET environment variable not set")
	}
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_REFRESH_SECRET")
	if secret == "" {
		return "", errors.New("JWT_REFRESH_SECRET environment variable not set")
	}
	return token.SignedString([]byte(secret))
}

// ─── Customer Portal Tokens ───────────────────────────────────────────────────
// Customer tokens use separate env vars (CUSTOMER_JWT_SECRET / CUSTOMER_JWT_REFRESH_SECRET)
// and carry {"sub": customerUUID, "type": "customer"} to prevent admin token reuse.

func GenerateCustomerAccessToken(customerID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  customerID,
		"type": "customer",
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("CUSTOMER_JWT_SECRET")
	if secret == "" {
		return "", errors.New("CUSTOMER_JWT_SECRET environment variable not set")
	}
	return token.SignedString([]byte(secret))
}

func GenerateCustomerRefreshToken(customerID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  customerID,
		"type": "customer",
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("CUSTOMER_JWT_REFRESH_SECRET")
	if secret == "" {
		return "", errors.New("CUSTOMER_JWT_REFRESH_SECRET environment variable not set")
	}
	return token.SignedString([]byte(secret))
}

func ValidateCustomerToken(tokenString string, isRefresh bool) (jwt.MapClaims, error) {
	secret := os.Getenv("CUSTOMER_JWT_SECRET")
	if isRefresh {
		secret = os.Getenv("CUSTOMER_JWT_REFRESH_SECRET")
	}
	if secret == "" {
		return nil, errors.New("customer JWT secret not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid customer token")
	}

	// Ensure this is a customer token (prevent admin token reuse)
	tokenType, _ := claims["type"].(string)
	if tokenType != "customer" {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}

func ValidateToken(tokenString string, isRefresh bool) (jwt.MapClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if isRefresh {
		secret = os.Getenv("JWT_REFRESH_SECRET")
	}
	if secret == "" {
		return nil, errors.New("JWT secret not set")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
