package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	stdprometheus "github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Make Http Handler
func MakeTcHttpHandler(ctx context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/taskManager/getTasks").Handler(httptransport.NewServer(
		endpoint.TaskCalendarEnpoint,
		decodeGetTasksRequest,
		encodeGetTasksResponce,
		append(options, httptransport.ServerBefore(jwt.HTTPToContext()))..., //auth
	))

	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		endpoint.HealthChecks,
		decodeHealthRequest,
		encodeHealthResponce,
		options...,
	))

	r.Path("/metrics").Handler(stdprometheus.Handler())

	return r
}

func decodeGetTasksRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request tasksRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeGetTasksResponce(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	var responce = resp.(tasksResponce)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(&responce); err != nil {
		return err
	}
	return nil
}

func decodeHealthRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	//var req healthRequest
	//if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	//	return nil, err
	//}
	//return req, nil
	return healthRequest{}, nil
}

func encodeHealthResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) (e error) {
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
