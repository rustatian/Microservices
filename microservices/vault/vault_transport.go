package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	stdprometheus "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Make Http Handler
func MakeVaultHttpHandler(svc Service) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		//httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/hash").Handler(httptransport.NewServer(
		makeHashEndpoint(svc),
		decodeHTTPHashRequest,
		encodeHTTPHashResponse,
		options...,
	))

	r.Methods("POST").Path("/validate").Handler(httptransport.NewServer(
		makeValidateEndpoint(svc),
		decodeHTTPValidateRequest,
		encodeHTTPValidateResponse,
		options...,
	))

	//GET /health
	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		makeHealthEndpoint(svc),
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
