package validator

import (
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidRepoFormat  = errors.New("invalid repo format")
	ErrInvalidTokenFormat = errors.New("invalid token")
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrInvalidEmail
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	return nil
}

func ParseAndValidateRepo(repo string) (string, string, error) {
	repo = strings.TrimSpace(repo)
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", ErrInvalidRepoFormat
	}

	owner := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])

	if owner == "" || name == "" {
		return "", "", ErrInvalidRepoFormat
	}

	return owner, name, nil
}

func ValidateToken(token string) error {
	token = strings.TrimSpace(token)

	if token == "" {
		return ErrInvalidTokenFormat
	}

	if len(token) != 64 {
		return ErrInvalidTokenFormat
	}

	_, err := hex.DecodeString(token)
	if err != nil {
		return ErrInvalidTokenFormat
	}

	return nil
}
