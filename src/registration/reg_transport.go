package registration

import (
	"context"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"

	"encoding/json"
)

//TODO replace nils
func MakeRegHttpHandler(_ context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}
	r.Methods("POST").Path("/registration").Handler(httptransport.NewServer(
		endpoint.RegEndpoint,
		decodeRegRequest,
		encodeRegResponce,
		options...,
	))

	r.Methods("POST").Path("/validate/user").Handler(httptransport.NewServer(
		endpoint.UsernameValidEndpoint,
		nil,
		nil,
		options...,
	))

	r.Methods("POST").Path("/validate/email").Handler(httptransport.NewServer(
		endpoint.EmailValidEndpoint,
		nil,
		nil,
		options...,
	))

	r.Methods("GET").Path("/validate/health").Handler(httptransport.NewServer(
		endpoint.RegHealthCheckEnpoint,
		nil,
		nil,
		options...,
	))

	return r
}

func decodeRegRequest(ctx context.Context, r *http.Request) (request interface{}, e error) {
	var req RegRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	req.isDisabled = false

	return req, nil
}

func encodeRegResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) error {
	if e, ok := responce.(error); ok && e != nil {
		return e
	}
	if _, ok := responce.(RegResponce); ok {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(responce)
}

func decodeUsernameValRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return "", nil
}

func encodeResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) error {
	return nil
}

func decodeEmailValidationRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return "", nil
}

func encodeEmailValidationResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) error {
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic("encodeError")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
