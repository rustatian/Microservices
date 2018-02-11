package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ValeryPiashchynski/TaskManager/microservices/vault/nats"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	gonats "github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Make Http Handler
func MakeVaultHttpHandler(endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	//r.Methods("POST").Path("/hash").Handler(httptransport.NewServer(
	//	endpoint.HashEnpoint,
	//	DecodeHashRequest,
	//	EncodeHashResponce,
	//	options...,
	//))

	r.Methods("POST").Path("/validate").Handler(httptransport.NewServer(
		endpoint.ValidateEndpoint,
		DecodeValidateRequest,
		EncodeValidateResponce,
		options...,
	))

	//GET /health
	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		endpoint.VaultHealtEndpoint,
		DecodeHealthRequest,
		EncodeHealthResponce,
		options...,
	))

	r.Path("/metrics").Handler(stdprometheus.Handler())

	handler := nats.NewServer(
		endpoint.HashNatsEnpoint,
		decodeUppercaseRequest,
		encodeResponse,
		5,
		10,
		5,
		time.Millisecond*10,
		nil,
	)

	nc, _ := gonats.Connect(gonats.DefaultURL)
	nc.QueueSubscribe("111", "111", handler.MsgHandler)

	return r
}

func decodeUppercaseRequest(_ context.Context, msg *gonats.Msg) (interface{}, error) {
	var request hashRequest
	if err := json.Unmarshal(msg.Data, &request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, response interface{}) (r []byte, err error) {
	resp := response.(hashResponse)
	data, err := json.Marshal(resp)
	return data, err
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
