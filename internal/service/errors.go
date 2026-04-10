package service

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidRepoFormat  = errors.New("invalid repo format")
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrAlreadySubscribed  = errors.New("already subscribed")
	ErrGitHubRateLimited  = errors.New("github rate limit exceeded")

	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenNotFound = errors.New("token not found")
)
