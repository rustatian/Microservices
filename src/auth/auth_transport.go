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
		encodeResponse,
		options...,
	))

	//auth logout
	r.Methods("POST").Path("/auth/logout").Handler(httptransport.NewServer(
		endpoint.LogoutEnpoint,
		decodeLogoutRequest,
		encodeResponse,
		options...,
	))

	//GET /health
	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		endpoint.HealthEndpoint,
		decodeHealthRequest,
		encodeResponse,
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
	return "", nil
}

func decodeHealthRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return HealthRequest{}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	type err interface {
		error() error
	}

	if e, ok := response.(err); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}

	if authResp, ok := response.(LoginResponce); ok {
		w.Header().Set("X-TOKEN-GEN", authResp.TokenString)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err == InvalidLoginErr {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}


