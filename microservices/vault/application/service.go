package application

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Hash(ctx context.Context, password string) (string, error)
}

type Validator interface {
	Validate(ctx context.Context, password, hash string) (bool, error)
}

type HealthChecker interface {
	HealthCheck() bool
}

type hasher struct{}

func (p *hasher) Hash(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

type validator struct{}

func (p *validator) Validate(ctx context.Context, password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}

type healthChecker struct{}

func (p *healthChecker) HealthCheck() bool {
	return true
}

func NewBcryptHasher() Hasher {
	return &hasher{}
}

func NewBcryptValidator() Validator {
	return &validator{}
}

func NewHttpHealthChecker() HealthChecker {
	return &healthChecker{}
}
