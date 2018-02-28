package tools

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type PasswordChecker interface {
	Hash(ctx context.Context, password string) (string, error)
	Validate(ctx context.Context, password, hash string) (bool, error)
	HealthCheck() bool
}

type passChecker struct {
}

func (p *passChecker) Hash(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (p *passChecker) Validate(ctx context.Context, password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *passChecker) HealthCheck() bool {
	return true
}

func NewPasswordChecker() PasswordChecker {
	return &passChecker{}
}
