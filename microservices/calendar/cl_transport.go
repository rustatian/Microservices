package calendar

import (
	"context"
	"encoding/json"
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

	//r.Methods("POST").Path("/validate").Handler(httptransport.NewServer(
	//	endpoint.ValidateEndpoint,
	//	DecodeValidateRequest,
	//	EncodeValidateResponce,
	//	options...,
	//))
	//
	////GET /health
	//r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
	//	endpoint.VaultHealtEndpoint,
	//	DecodeHealthRequest,
	//	EncodeHealthResponce,
	//	options...,
	//))

	r.Path("/metrics").Handler(stdprometheus.Handler())

	return r
}

func decodeGetTasksRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request TasksRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeGetTasksResponce(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	var responce = resp.(TasksResponce)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(&responce); err != nil {
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
