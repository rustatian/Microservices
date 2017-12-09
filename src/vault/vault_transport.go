package vault

import (
	httptransport "github.com/go-kit/kit/transport/http"
	"context"
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"github.com/go-kit/kit/sd"
	"io"
	"strings"
	"net/url"
	"github.com/go-kit/kit/endpoint"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/go-kit/kit/log"
)

// Make Http Handler
func MakeHttpHandler(_ context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}


	r.Methods("POST").Path("/vaultsvc/hash").Handler(httptransport.NewServer(
		endpoint.HashEnpoint,
		DecodeHashRequest,
		EncodeResponce,
		options...,
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

	// GET /metrics
	//r.Path("/metrics").Handler(stdprometheus.Handler())

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

func VaultSvcFactory(ctx context.Context, method, path string) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}
		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		tgt.Path = path

		var (
			enc httptransport.EncodeRequestFunc
			dec httptransport.DecodeResponseFunc
		)
		switch path {
		case "/hash":
			enc, dec = EncodeRequest, DecodeHashResponce
		case "/validate":
			enc, dec = EncodeRequest, DecodeValidateResponce
		case "/health":
			enc, dec = EncodeRequest, DecodeHealthResponce
		default:
			return nil, nil, fmt.Errorf("unknown stringsvc path %q", path)
		}

		return httptransport.NewClient(method, tgt, enc, dec).Endpoint(), nil, nil
	}
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