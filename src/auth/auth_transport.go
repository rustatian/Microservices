package auth

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

// Make Http Handler
func MakeAuthHttpHandler(_ context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	//auth login
	r.Methods("POST").Path("/auth/login").Handler(httptransport.NewServer(
		endpoint.LoginEndpoint,
		decodeLoginRequest,
		encodeLoginResponse,
		options...,
	))

	//auth logout
	r.Methods("POST").Path("/auth/logout").Handler(httptransport.NewServer(
		endpoint.LogoutEnpoint,
		decodeLogoutRequest,
		encodeLogoutResponce,
		options...,
	))

	//GET /health
	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		endpoint.HealthEndpoint,
		decodeHealthRequest,
		encodeLoginResponse,
		options...,
	))
	return r
}

func decodeLoginRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	val := r.Header.Get("Authorization")
	loginHeaderParts := strings.Split(val, " ")

	//Check if there is - bearer wqeorij384u2-384u9
	if len(loginHeaderParts) == 2 && strings.ToLower(loginHeaderParts[0]) == "bearer" {
		req.TokenString = loginHeaderParts[1]
	}
	return req, nil
}

func decodeLogoutRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", err
	}

	return req, nil
}

func decodeHealthRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return HealthRequest{}, nil
}

func encodeLoginResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(error); ok && e != nil {
		encodeError(ctx, e, w)
		return nil
	}

	if authResp, ok := response.(LoginResponce); ok {
		w.Header().Set("X-TOKEN-GEN", authResp.TokenString)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeLogoutResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) error {
	if e, ok := responce.(error); ok && e != nil {
		encodeError(ctx, e, w)
	}

	return json.NewEncoder(w).Encode(responce)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}





















