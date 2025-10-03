package utils

import (
	"errors"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ExtractUserIDFromHeader parses Authorization header (Bearer <token>) and returns user_id UUID from JWT claims.
func ExtractUserIDFromHeader(authHeader string) (uuid.UUID, error) {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return uuid.Nil, errors.New("missing or invalid Authorization header")
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return uuid.Nil, errors.New("JWT secret not set")
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) { return []byte(secret), nil })
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid token payload")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		println(err.Error())
		return uuid.Nil, errors.New("invalid user id in token")
	}
	return userID, nil
}
