package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/ValeryPiashchynski/TaskManager/microservices/vault/nats"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	gonats "github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func MakeVaultNatsHandler(endpoint Endpoints, logger log.Logger, natsUrl string, err chan error) http.Handler {
	//Used for health checks and metrics
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	nc, er := gonats.Connect(natsUrl)
	if er != nil {
		err <- er
	}
	nc, er = gonats.Connect(natsUrl,
		gonats.DisconnectHandler(func(nc *gonats.Conn) {
			fmt.Printf("Got disconnected!\n")
		}),
		gonats.ReconnectHandler(func(_ *gonats.Conn) {
			fmt.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
		}),
		gonats.ClosedHandler(func(nc *gonats.Conn) {
			fmt.Printf("Connection closed. Reason: %q\n", nc.LastError())
		}),
	)

	if er != nil {
		err <- er
	}

	hashHandler := nats.NewServer(
		endpoint.HashEndpoint,
		decodeHashRequest,
		encodeHashResponse,
		10,
		runtime.GOMAXPROCS(12),
		10,
		time.Microsecond*1,
		err,
	)
	nc.QueueSubscribe("hash", "hash", hashHandler.MsgHandler)

	validateHandler := nats.NewServer(
		endpoint.ValidateEndpoint,
		decodeValidateRequest,
		encodeValidateResponse,
		10,
		runtime.GOMAXPROCS(12),
		10,
		time.Microsecond*1,
		err,
	)
	nc.QueueSubscribe("validate", "validate", validateHandler.MsgHandler)

	//GET /health
	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		endpoint.VaultHealthEndpoint,
		decodeHTTPHealthRequest,
		encodeHTTPHealthResponse,
		options...,
	))
	r.Path("/metrics").Handler(stdprometheus.Handler())

	return r
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
		decodeHTTPHashRequest,
		encodeHTTPHashResponse,
		options...,
	))

	r.Methods("POST").Path("/validate").Handler(httptransport.NewServer(
		endpoint.ValidateEndpoint,
		decodeHTTPValidateRequest,
		encodeHTTPValidateResponse,
		options...,
	))

	//GET /health
	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		endpoint.VaultHealthEndpoint,
		decodeHTTPHealthRequest,
		encodeHTTPHealthResponse,
		options...,
	))

	r.Path("/metrics").Handler(stdprometheus.Handler())

	return r
}

func decodeHTTPHashRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request hashRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeHTTPValidateRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req validateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeHTTPHealthRequest(ctx context.Context, r *http.Request) (interface{}, error) {

	contx, err := GetContext(r)

	if err != nil {
		contx.Log.WithFields(logrus.Fields{
			"Error":   err.Error(),
			"request": r,
		}).Error("Decode health request error")
	}

	contx.Log.WithFields(logrus.Fields{
		"time":    time.Now().Format(time.RFC3339Nano),
		"Method":  "decodeHTTPHealthRequest",
		"request": r,
	}).Info("Decode health request")
	return healthRequest{}, nil
}

func encodeHTTPHashResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	var responce = resp.(hashResponse)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(&responce); err != nil {
		return err
	}
	return nil
}

func encodeHTTPValidateResponse(ctx context.Context, w http.ResponseWriter, responce interface{}) (e error) {
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

func encodeHTTPHealthResponse(ctx context.Context, w http.ResponseWriter, responce interface{}) (e error) {
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

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
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
