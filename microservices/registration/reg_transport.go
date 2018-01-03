package registration

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
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

	r.Methods("POST").Path("/registration/user").Handler(httptransport.NewServer(
		endpoint.UsernameValidEndpoint,
		decodeUsernameValRequest,
		encodeUsernameValResponce,
		options...,
	))

	r.Methods("POST").Path("/registration/email").Handler(httptransport.NewServer(
		endpoint.EmailValidEndpoint,
		decodeEmailValidationRequest,
		encodeEmailValidationResponce,
		options...,
	))

	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		endpoint.RegHealthCheckEnpoint,
		decodeRegHealthCheckRequest,
		encodeRegHealthCheckResponce,
		options...,
	))

	r.Path("/metrics").Handler(promhttp.Handler())

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
	if resp, ok := responce.(RegResponce); ok {
		if resp.Status == true {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(resp)
			return nil
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return nil

	} else {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(ok)
	}
}

func decodeUsernameValRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req UsernameValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func encodeUsernameValResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) error {
	if resp, ok := responce.(UsernameValidationResponce); ok {
		if resp.Status == true {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(resp)
			return nil
		}

		w.WriteHeader(http.StatusOK)
		//w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(resp)
		return nil
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(ok)
	}
}

func decodeEmailValidationRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req EmailValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func encodeEmailValidationResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) error {
	if resp, ok := responce.(EmailValidationResponce); ok {
		if resp.Status == true {
			w.WriteHeader(http.StatusConflict)
			//w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(resp)
		}
		w.WriteHeader(http.StatusOK)
		//w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(resp)
		return nil
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(ok)
	}
}

func decodeRegHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	//var req HealthRequest
	//if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	//	return nil, err
	//}
	//return req, nil
	return HealthRequest{}, nil
}

func encodeRegHealthCheckResponce(ctx context.Context, w http.ResponseWriter, responce interface{}) error {
	if resp, ok := responce.(HealthResponse); ok {
		w.WriteHeader(http.StatusOK)
		//w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(resp)
		return nil
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(ok)
	}
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
