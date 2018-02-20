package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ValeryPiashchynski/TaskManager/microservices/pb/vault"
	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	stdprometheus "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type grpcServer struct {
	hash grpctransport.Handler
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

func NewGrpcServer(ctx context.Context, endpoints Endpoints) pb_vault.VaultServer {
	return &grpcServer{
		hash: grpctransport.NewServer(
			endpoints.HashEndpoint,
			decodeGRPCHashRequest,
			encodeGRPCHashResponce,
		),
	}

}

// Make Http Handler
func MakeVaultHttpHandler(endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/hash").Handler(httptransport.NewServer(
		endpoint.HashEndpoint,
		DecodeHashRequest,
		EncodeHashResponce,
		options...,
	))

	r.Methods("POST").Path("/validate").Handler(httptransport.NewServer(
		endpoint.ValidateEndpoint,
		DecodeValidateRequest,
		EncodeValidateResponce,
		options...,
	))

	//GET /health
	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		endpoint.VaultHealthEndpoint,
		DecodeHealthRequest,
		EncodeHealthResponce,
		options...,
	))

	r.Path("/metrics").Handler(stdprometheus.Handler())

	return r
}

//Encode + Decode hashRequest
func encodeGRPCHashRequest(_ context.Context, request interface{}) (response interface{}, err error) {
	req := request.(hashRequest)
	return &pb_vault.HashRequest{
		Password: req.Password,
	}, nil

}

func decodeGRPCHashRequest(_ context.Context, request interface{}) (response interface{}, err error) {
	req := request.(*pb_vault.HashRequest)
	return hashRequest{
		Password: req.Password,
	}, nil
}

func encodeGRPCHashResponce(_ context.Context, request interface{}) (response interface{}, err error) {
	req := request.(hashResponse)
	return &pb_vault.HashResponce{
		Hash: req.Hash,
		Err:  req.Err,
	}, nil
}

func decodeGRPCHashResponce(_ context.Context, request interface{}) (response interface{}, err error) {
	req := request.(*pb_vault.HashResponce)
	return hashResponse{
		Hash: req.Hash,
		Err:  req.Err,
	}, nil
}

func DecodeHashRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request hashRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func DecodeValidateRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req validateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func DecodeHealthRequest(ctx context.Context, r *http.Request) (interface{}, error) {

	contx, err := GetContext(r)

	if err != nil {
		contx.Log.WithFields(logrus.Fields{
			"Error":   err.Error(),
			"request": r,
		}).Error("Decode health request error")
	}

	contx.Log.WithFields(logrus.Fields{
		"time":    time.Now().Format(time.RFC3339Nano),
		"Method":  "DecodeHealthRequest",
		"request": r,
	}).Info("Decode health request")

	//var req healthRequest
	//if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	//	return nil, err
	//}
	//return req, nil
	return healthRequest{}, nil
}

func EncodeHashResponce(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	var responce = resp.(hashResponse)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(&responce); err != nil {
		return err
	}
	return nil
}

func EncodeValidateResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) (e error) {
	resp, ok := responce.(validateResponse)
	if !ok {
		return fmt.Errorf("type conversion error in validate encode responce")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return err
	}

	return nil
}

func EncodeHealthResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) (e error) {
	resp, ok := responce.(healthResponse)
	if !ok {
		return fmt.Errorf("type conversion error in health encode responce")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		return err
	}

	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
