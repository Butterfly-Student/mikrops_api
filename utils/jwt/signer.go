package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type StaffClaims struct {
	StaffID     string   `json:"staff_id"`
	TenantID    string   `json:"tenant_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

type CustomerClaims struct {
	CustomerID string `json:"customer_id"`
	TenantID   string `json:"tenant_id"`
	jwt.RegisteredClaims
}

func GenerateStaffToken(staffID, tenantID, role string, permissions []string, secret string) (string, error) {
	claims := StaffClaims{
		StaffID:     staffID,
		TenantID:    tenantID,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mikrops",
			Subject:   staffID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateStaffRefreshToken(staffID, tenantID string, secret string) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "mikrops",
		Subject:   staffID,
		ID:        tenantID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateCustomerToken(customerID, tenantID, secret string) (string, error) {
	claims := CustomerClaims{
		CustomerID: customerID,
		TenantID:   tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mikrops",
			Subject:   customerID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateCustomerRefreshToken(customerID, tenantID, secret string) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "mikrops",
		Subject:   customerID,
		ID:        tenantID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseStaffToken(tokenString, secret string) (*StaffClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &StaffClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse staff token: %w", err)
	}

	claims, ok := token.Claims.(*StaffClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid staff token")
	}

	return claims, nil
}

func ParseCustomerToken(tokenString, secret string) (*CustomerClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomerClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse customer token: %w", err)
	}

	claims, ok := token.Claims.(*CustomerClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid customer token")
	}

	return claims, nil
}

func ParseRefreshToken(tokenString, secret string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	return claims, nil
}
