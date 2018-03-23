package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	gknats "github.com/ValeryPiashchynski/TaskManager/microservices/tools/nats"
	customhttptransport "github.com/ValeryPiashchynski/TaskManager/microservices/vault/infrastructure"
	"github.com/gorilla/mux"
	gonats "github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func StartVaultNatsHandler(endpoint Endpoints, logger logrus.Logger, stopSignal chan os.Signal, conn *gonats.Conn, err chan error) http.Handler {
	r := mux.NewRouter()
	options := []customhttptransport.ServerOption{
		customhttptransport.ServerErrorLogger(logger),
		customhttptransport.ServerErrorEncoder(encodeError),
	}

	hashHandler := gknats.NewServer(
		endpoint.HashEndpoint,
		decodeHashRequest,
		encodeHashResponse,
		conn,
		logger,
		stopSignal,
		2,
		"hash",
		err,
	)
	conn.QueueSubscribe("hash", "hash", hashHandler.MsgHandler)

	validateHandler := gknats.NewServer(
		endpoint.ValidateEndpoint,
		decodeValidateRequest,
		encodeValidateResponse,
		conn,
		logger,
		stopSignal,
		2,
		"hash",
		err,
	)
	conn.QueueSubscribe("validate", "validate", validateHandler.MsgHandler)

	//GET /health
	r.Methods("GET").Path("/health").Handler(customhttptransport.NewServer(
		endpoint.HealthCheckEndpoint,
		decodeHTTPHealthRequest,
		encodeHTTPHealthResponse,
		options...,
	))
	r.Path("/metrics").Handler(stdprometheus.Handler())

	return r
}

//NATS Encode/decode
func encodeHashResponse(ctx context.Context, response interface{}) (r []byte, err error) {
	resp := response.(hashResponse)
	data, err := json.Marshal(resp)
	return data, err
}

func decodeHashRequest(ctx context.Context, msg *gonats.Msg) (interface{}, error) {
	var request hashRequest
	if err := json.Unmarshal(msg.Data, &request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeValidateResponse(ctx context.Context, response interface{}) (r []byte, err error) {
	resp := response.(validateResponse)
	data, err := json.Marshal(resp)
	return data, err
}

func decodeValidateRequest(ctx context.Context, msg *gonats.Msg) (interface{}, error) {
	var request validateRequest
	if err := json.Unmarshal(msg.Data, &request); err != nil {
		return nil, err
	}
	return request, nil
}
