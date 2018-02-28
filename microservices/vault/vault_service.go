package vault

import (
	"context"
	"errors"
	"github.com/ValeryPiashchynski/TaskManager/microservices/tools"
	"github.com/go-kit/kit/endpoint"
)

type Service interface {
	Hash(ctx context.Context, password string) (string, error)
	Validate(ctx context.Context, password, hash string) (bool, error)
	HealthCheck() bool
}

func NewVaultService(checker tools.PasswordChecker) Service {
	return &service{
		pswChecker: checker,
	}
}

type ServiceMiddleware func(svc Service) Service

type service struct {
	pswChecker tools.PasswordChecker
}

func (s *service) Hash(ctx context.Context, password string) (string, error) {
	hash, err := s.pswChecker.Hash(ctx, password)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *service) Validate(ctx context.Context, password, hash string) (bool, error) {
	ok, err := s.pswChecker.Validate(ctx, password, hash)
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (s *service) HealthCheck() bool {
	return s.pswChecker.HealthCheck()
}

type Endpoints struct {
	HashEndpoint        endpoint.Endpoint
	ValidateEndpoint    endpoint.Endpoint
	VaultHealthEndpoint endpoint.Endpoint
}

func (e Endpoints) Hash(ctx context.Context, password string) (string, error) {
	req := hashRequest{Password: password}
	resp, err := e.HashEndpoint(ctx, req)
	if err != nil {
		return "", err
	}
	hashResp := resp.(hashResponse)
	if hashResp.Err != "" {
		return "", errors.New(hashResp.Err)
	}
	return hashResp.Hash, nil
}

func (e Endpoints) Validate(ctx context.Context, password, hash string) (bool, error) {
	req := validateRequest{Password: password, Hash: hash}
	resp, err := e.ValidateEndpoint(ctx, req)
	if err != nil {
		return false, err
	}
	validateResp := resp.(validateResponse)
	if validateResp.Err != "" {
		return false, errors.New(validateResp.Err)
	}
	return validateResp.Valid, nil
}

func MakeHashEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(hashRequest)
		v, err := svc.Hash(ctx, req.Password)
		if err != nil {
			return hashResponse{Hash: v, Err: err.Error()}, err
		}
		return hashResponse{Hash: v, Err: ""}, nil
	}
}

func MakeValidateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(validateRequest)
		v, err := svc.Validate(ctx, req.Password, req.Hash)
		if err != nil {
			return validateResponse{Valid: v, Err: err.Error()}, err
		}
		return validateResponse{Valid: v, Err: ""}, nil
	}
}

func MakeHealthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		//req := request.(healthRequest)
		v := svc.HealthCheck()
		return healthResponse{Status: v}, nil
	}
}
