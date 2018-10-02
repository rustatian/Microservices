package vault

import (
	"Microservices/microservices/vault/application"
	"context"
)

type Service interface {
	Hash(ctx context.Context, password string) (string, error)
	Validate(ctx context.Context, password, hash string) (bool, error)
	HealthCheck() bool
}

// NewVaultService - constructor for vault service
func NewVaultService(hasher application.Hasher, validator application.Validator, checker application.HealthChecker) Service {
	return &service{
		hash:          hasher,
		validate:      validator,
		healthChecker: checker,
	}
}

type ServiceMiddleware func(svc Service) Service

type service struct {
	hash          application.Hasher
	validate      application.Validator
	healthChecker application.HealthChecker
}

func (s *service) Hash(ctx context.Context, password string) (string, error) {
	hash, err := s.hash.Hash(ctx, password)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *service) Validate(ctx context.Context, password, hash string) (bool, error) {
	ok, err := s.validate.Validate(ctx, password, hash)
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (s *service) HealthCheck() bool {
	return s.healthChecker.HealthCheck()
}
