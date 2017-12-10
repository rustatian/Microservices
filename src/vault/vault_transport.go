package vault

import (
	httptransport "github.com/go-kit/kit/transport/http"
	"context"
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"github.com/gorilla/mux"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/auth/jwt"
)

// Make Http Handler
func MakeVaultHttpHandler(_ context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/vaultsvc/hash").Handler(httptransport.NewServer(
		endpoint.HashEnpoint,
		DecodeHashRequest,
		EncodeResponce,
		append(options, httptransport.ServerBefore(jwt.HTTPToContext()))...,
	))

	r.Methods("POST").Path("/vaultsvc/validate").Handler(httptransport.NewServer(
		endpoint.ValidateEndpoint,
		DecodeValidateRequest,
		EncodeResponce,
		options...,
	))

	//GET /health
	r.Methods("GET").Path("/vaultsvc/health").Handler(httptransport.NewServer(
		endpoint.VaultHealtEndpoint,
		DecodeHealthRequest,
		EncodeResponce,
		options...,
	))

	return r
}

func DecodeHashRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request hashRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func DecodeValidateRequest(ctx context.Context, r *http.Request) (interface{}, error)  {
	var req validateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func DecodeHealthRequest(ctx context.Context, r *http.Request) (interface{}, error)  {
	//var req healthRequest
	//if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	//	return nil, err
	//}
	//return req, nil
	return healthRequest{}, nil
}


func DecodeHashResponce(ctx context.Context, r *http.Response) (interface{}, error) {
	var responce hashResponse
	if err := json.NewDecoder(r.Body).Decode(&responce); err != nil {
		return nil, err
	}
	return responce, nil
}

func DecodeValidateResponce(ctx context.Context, r *http.Response) (interface{}, error) {
	var responce validateResponse
	if err := json.NewDecoder(r.Body).Decode(&responce); err != nil {
		return nil, err
	}
	return responce, nil
}

func DecodeHealthResponce(ctx context.Context, r *http.Response) (interface{}, error) {
	var responce healthResponse
	if err := json.NewDecoder(r.Body).Decode(&responce); err != nil {
		return nil, err
	}
	return responce, nil
}


func EncodeResponce(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func EncodeRequest(_ context.Context, req *http.Request, request interface{}) error {
	// Both uppercase and count requests are encoded in the same way:
	// simple JSON serialization to the request body.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
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