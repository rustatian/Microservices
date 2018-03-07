package vault

import (
	"context"
)

type Service interface {
	Hash(ctx context.Context, password string) (string, error)
	Validate(ctx context.Context, password, hash string) (bool, error)
	HealthCheck() bool
}

func NewVaultService(checker Service) Service {
	return &service{
		hasher: checker,
	}
}

type ServiceMiddleware func(svc Service) Service

type service struct {
	hasher Service
}

func (s *service) Hash(ctx context.Context, password string) (string, error) {
	hash, err := s.hasher.Hash(ctx, password)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *service) Validate(ctx context.Context, password, hash string) (bool, error) {
	ok, err := s.hasher.Validate(ctx, password, hash)
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (s *service) HealthCheck() bool {
	return s.hasher.HealthCheck()
}
