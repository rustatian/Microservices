package vault

import (
	"github.com/ValeryPiashchynski/TaskManager/microservices/proto/vault"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"golang.org/x/net/context"
)

type grpcServer struct {
	hash     grpctransport.Handler
	validate grpctransport.Handler
	health   grpctransport.Handler
}

func (g *grpcServer) Hash(ctx context.Context, r *pb_vault.HashRequest) (*pb_vault.HashResponce, error) {
	_, resp, err := g.hash.ServeGRPC(ctx, r)
	if err != nil {
		return nil, err
	}
	return resp.(*pb_vault.HashResponce), nil
}

func (g *grpcServer) Validate(ctx context.Context, r *pb_vault.ValidateRequest) (*pb_vault.ValidateResponce, error) {
	_, resp, err := g.hash.ServeGRPC(ctx, r)
	if err != nil {
		return nil, err
	}
	return resp.(*pb_vault.ValidateResponce), nil
}

func (g *grpcServer) HealthCheck(ctx context.Context, r *pb_vault.HealthRequest) (*pb_vault.HealthResponse, error) {
	_, resp, err := g.hash.ServeGRPC(ctx, r)
	if err != nil {
		return nil, err
	}
	return resp.(*pb_vault.HealthResponse), nil
}
