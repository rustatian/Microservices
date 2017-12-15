package registration

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"net/http"
	httptransport "github.com/go-kit/kit/transport/http"

	"encoding/json"
)

//TODO replace nils
func MakeRegHandler(_ context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}
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


	return r
}

func decodeUsernameValRequest(ctx context.Context, r *http.Request) (interface{}, error) {

}

func encodeResponce


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
