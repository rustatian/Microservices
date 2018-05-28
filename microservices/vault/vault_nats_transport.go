package vault

import (
	"context"
	"encoding/json"
	"net/http"

	gknats "github.com/ValeryPiashchynski/Microservices/microservices/tools/nats"
	customhttptransport "github.com/ValeryPiashchynski/Microservices/microservices/vault/infrastructure"
	"github.com/gorilla/mux"
	gonats "github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func StartVaultNatsHandler(endpoint Endpoints, logger logrus.Logger, conn *gonats.Conn) http.Handler {
	r := mux.NewRouter()
	options := []customhttptransport.ServerOption{
		customhttptransport.ServerErrorLogger(logger),
		customhttptransport.ServerErrorEncoder(encodeError),
	}

	hashHandler := gknats.NewSubscriber(
		endpoint.HashEndpoint,
		decodeHashRequest,
		encodeHashResponse,
	)

	_, err := conn.QueueSubscribe("hash", "hash", hashHandler.ServeMsg(conn))
	if err != nil {
		logrus.Fatal(err)
	}
	//defer hsub.Unsubscribe()

	validateHandler := gknats.NewSubscriber(
		endpoint.ValidateEndpoint,
		decodeValidateRequest,
		encodeValidateResponse,
	)

	_, err = conn.QueueSubscribe("validate", "validate", validateHandler.ServeMsg(conn))
	if err != nil {
		logrus.Fatal(err)
	}
	//defer vsub.Unsubscribe()

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
func encodeHashResponse(ctx context.Context, reply string, nc *gonats.Conn, response interface{}) (err error) {
	resp := response.(hashResponse)
	data, err := json.Marshal(resp)
	return nc.Publish(reply, data)
}

func decodeHashRequest(ctx context.Context, msg *gonats.Msg) (interface{}, error) {
	var request hashRequest
	if err := json.Unmarshal(msg.Data, &request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeValidateResponse(ctx context.Context, reply string, nc *gonats.Conn, response interface{}) (err error) {
	resp := response.(validateResponse)
	data, err := json.Marshal(resp)
	return nc.Publish(reply, data)
}

func decodeValidateRequest(ctx context.Context, msg *gonats.Msg) (interface{}, error) {
	var request validateRequest
	if err := json.Unmarshal(msg.Data, &request); err != nil {
		return nil, err
	}
	return request, nil
}
