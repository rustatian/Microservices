package vault

import (
	"context"
	"fmt"

	"github.com/ValeryPiashchynski/TaskManager/microservices/proto/vault"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

func MakeVaultGrpcHandler(svc Service) pb_vault.VaultServer {
	options := []grpctransport.ServerOption{
		//httptransport.ServerErrorLogger(logger),
		//grpctransport.ServerErrorEncoder(encodeError),
	}

	return &grpcServer{
		hash: grpctransport.NewServer(
			makeHashEndpoint(svc),
			decodeGRPCHashRequest,
			encodeGRPCHashResponse,
			options...,
		),
		validate: grpctransport.NewServer(
			makeValidateEndpoint(svc),
			decodeGRPCValidateRequest,
			encodeGRPCValidateResponse,
			options...,
		),
		health: grpctransport.NewServer(
			makeHealthEndpoint(svc),
			decodeGRPCHealthRequest,
			encodeGRPCHealthResponse,
			options...,
		),
	}
}

func decodeGRPCHashRequest(ctx context.Context, r interface{}) (interface{}, error) {
	req := r.(*pb_vault.HashRequest)
	return hashRequest{Password: req.Password}, nil
}

func encodeGRPCHashResponse(ctx context.Context, r interface{}) (interface{}, error) {
	res := r.(hashResponse)
	return &pb_vault.HashResponce{Hash: res.Hash, Err: res.Err.Error()},
		nil
}

func decodeGRPCHashResponse(ctx context.Context, r interface{}) (interface{}, error) {
	res := r.(*pb_vault.HashResponce)
	return hashResponse{Hash: res.Hash, Err: fmt.Errorf(res.Err)}, nil
}
func encodeGRPCValidateRequest(ctx context.Context, r interface{}) (interface{}, error) {
	req := r.(validateRequest)
	return &pb_vault.ValidateRequest{Password: req.Password,
		Hash: req.Hash}, nil
}

func decodeGRPCValidateRequest(ctx context.Context, r interface{}) (interface{}, error) {
	req := r.(*pb_vault.ValidateRequest)
	return validateRequest{Password: req.Password,
		Hash: req.Hash}, nil
}
func encodeGRPCValidateResponse(ctx context.Context, r interface{}) (interface{}, error) {
	res := r.(validateResponse)
	return &pb_vault.ValidateResponce{Valid: res.Valid}, nil
}

func decodeGRPCValidateResponse(ctx context.Context, r interface{}) (interface{}, error) {
	res := r.(*pb_vault.ValidateResponce)
	return validateResponse{Valid: res.Valid}, nil
}

func decodeGRPCHealthRequest(ctx context.Context, r interface{}) (interface{}, error) {
	//req := r.(*pb_vault.HealthRequest)
	return healthRequest{}, nil
}

func encodeGRPCHealthResponse(ctx context.Context, r interface{}) (interface{}, error) {
	//res := r.(healthResponse)
	return &pb_vault.HealthResponse{},
		nil
}
