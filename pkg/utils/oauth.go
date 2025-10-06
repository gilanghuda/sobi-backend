package utils

import (
	"context"
	"errors"
	"os"

	"cloud.google.com/go/auth/credentials/idtoken"
)

func ValidateGoogleIDToken(ctx context.Context, idToken string) (string, error) {
	audience := os.Getenv("OAUTH_CLIENT_ID")
	if audience == "" {
		return "", errors.New("oauth client id not configured (tidak ada)")
	}

	tok, err := idtoken.Validate(ctx, idToken, audience)
	if err != nil {
		return "", err
	}

	emailIF, ok := tok.Claims["email"]
	if !ok {
		return "", errors.New("google token does not contain email")
	}
	email, ok2 := emailIF.(string)
	if !ok2 || email == "" {
		return "", errors.New("invalid email claim in google token")
	}
	return email, nil
}
