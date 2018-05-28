package vault

import (
	"context"
	"fmt"

	"github.com/ValeryPiashchynski/Microservices/microservices/proto/vault"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

func MakeVaultGrpcHandler(svc Service) pb_vault.VaultServer {
	options := []grpctransport.ServerOption{}

	return &grpcServer{
		hash: grpctransport.NewServer(
			makeHashEndpoint(svc),
			DecodeGRPCHashRequest,
			EncodeGRPCHashResponse,
			options...,
		),
		validate: grpctransport.NewServer(
			makeValidateEndpoint(svc),
			DecodeGRPCValidateRequest,
			EncodeGRPCValidateResponse,
			options...,
		),
		health: grpctransport.NewServer(
			makeHealthEndpoint(svc),
			DecodeGRPCHealthRequest,
			EncodeGRPCHealthResponse,
			options...,
		),
	}
}

//Hash\\
func DecodeGRPCHashRequest(ctx context.Context, r interface{}) (interface{}, error) {
	req := r.(*pb_vault.HashRequest)
	return hashRequest{Password: req.Password}, nil
}
func DecodeGRPCHashResponse(ctx context.Context, r interface{}) (interface{}, error) {
	res := r.(*pb_vault.HashResponce)
	return hashResponse{Hash: res.Hash, Err: fmt.Errorf(res.Err)}, nil
}

func EncodeGRPCHashResponse(ctx context.Context, r interface{}) (interface{}, error) {
	res := r.(hashResponse)
	if res.Err != nil {
		return nil, res.Err
	}
	return &pb_vault.HashResponce{Hash: res.Hash, Err: ""}, nil
}
func EncodeGRPCHashRequest(ctx context.Context, r interface{}) (interface{}, error) {
	req := r.(hashRequest)
	return &pb_vault.HashRequest{Password: req.Password}, nil
}

// Validate \\
func DecodeGRPCValidateRequest(ctx context.Context, r interface{}) (interface{}, error) {
	req := r.(*pb_vault.ValidateRequest)
	return validateRequest{Password: req.Password,
		Hash: req.Hash}, nil
}
func DecodeGRPCValidateResponse(ctx context.Context, r interface{}) (interface{}, error) {
	res := r.(*pb_vault.ValidateResponce)
	return validateResponse{Valid: res.Valid}, nil
}
func EncodeGRPCValidateRequest(ctx context.Context, r interface{}) (interface{}, error) {
	req := r.(validateRequest)
	return &pb_vault.ValidateRequest{Password: req.Password, Hash: req.Hash}, nil
}
func EncodeGRPCValidateResponse(ctx context.Context, r interface{}) (interface{}, error) {
	res := r.(validateResponse)
	return &pb_vault.ValidateResponce{Valid: res.Valid}, nil
}

// Health \\
func DecodeGRPCHealthRequest(ctx context.Context, r interface{}) (interface{}, error) {
	//req := r.(*pb_vault.HealthRequest)
	return healthRequest{}, nil
}
func EncodeGRPCHealthResponse(ctx context.Context, r interface{}) (interface{}, error) {
	//res := r.(healthResponse)
	return &pb_vault.HealthResponse{},
		nil
}
