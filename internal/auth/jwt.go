package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Authorized bool   `json:"authorized"`
	UserID     string `json:"user_id"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int, role string, secret string) (string, time.Duration, error) {
	var tokenExp time.Duration
	if role == "admin" {
		tokenExp = 24 * time.Hour
	} else {
		tokenExp = 5 * time.Minute
	}

	userIDStr := strconv.Itoa(userID)

	claims := Claims{
		Authorized: true,
		UserID:     userIDStr,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Cinema-backend",
			Subject:   userIDStr,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}

	return signedToken, tokenExp, nil
}

func ValidateToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while parsing")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
