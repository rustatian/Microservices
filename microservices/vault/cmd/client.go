package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"Microservices/microservices/proto/vault"
	"Microservices/microservices/vault"

	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

func main() {
	var (
		grpcAddr = flag.String("addr", ":8081", "gRPC address")
	)
	flag.Parse()
	ctx := context.Background()
	conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(1*time.Second))
	if err != nil {
		log.Fatalln("gRPC dial:", err)
	}
	defer conn.Close()
	vaultService := New(conn)

	password := "12345678"
	hash(ctx, vaultService, password)
}

func hash(ctx context.Context, service vault.Service, password string) {
	h, err := service.Hash(ctx, password)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println(h)
}

func New(conn *grpc.ClientConn) vault.Service {
	var hashEndpoint = grpctransport.NewClient(
		conn, "pb.vault.Vault", "Hash",
		vault.EncodeGRPCHashRequest,
		vault.DecodeGRPCHashResponse,
		pb_vault.HashResponce{},
	).Endpoint()
	var validateEndpoint = grpctransport.NewClient(
		conn, "pb.vault.Vault", "Validate",
		vault.EncodeGRPCValidateRequest,
		vault.DecodeGRPCValidateResponse,
		pb_vault.ValidateResponce{},
	).Endpoint()

	return vault.Endpoints{
		HashEndpoint:     hashEndpoint,
		ValidateEndpoint: validateEndpoint,
	}
}
