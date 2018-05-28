package vault

import (
	"context"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type hashRequest struct {
	Password string `json:"password"`
}

type hashResponse struct {
	Hash string `json:"hash"`
	Err  error  `json:"error, omitempty"`
}

func (h hashResponse) error() error {
	return h.Err
}

func makeHashEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(hashRequest)
		v, err := svc.Hash(ctx, req.Password)
		return hashResponse{Hash: v, Err: err}, nil
	}
}

type validateRequest struct {
	Password string `json:"password"`
	Hash     string `json:"hash"`
}

type validateResponse struct {
	Valid bool  `json:"valid"`
	Err   error `json:"error, omitempty"`
}

func (v validateResponse) error() error {
	return v.Err
}

func makeValidateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(validateRequest)
		v, err := svc.Validate(ctx, req.Password, req.Hash)
		return validateResponse{Valid: v, Err: err}, nil
	}
}

//TODO Create health logic, check free memory, disk space
type healthRequest struct{}

type healthResponse struct {
	Status bool `json:"status"`
}

func makeHealthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		//req := request.(healthRequest)
		v := svc.HealthCheck()
		return healthResponse{Status: v}, nil
	}
}

type Endpoints struct {
	HashEndpoint        endpoint.Endpoint
	ValidateEndpoint    endpoint.Endpoint
	HealthCheckEndpoint endpoint.Endpoint
}

func (e Endpoints) Hash(ctx context.Context, password string) (string, error) {
	req := hashRequest{Password: password}
	resp, err := e.HashEndpoint(ctx, req)
	if err != nil {
		return "", err
	}
	hashResp := resp.(hashResponse)
	if hashResp.Err != nil {
		return "", hashResp.Err
	}
	return hashResp.Hash, nil
}

func (e Endpoints) Validate(ctx context.Context, password, hash string) (bool, error) {
	req := validateRequest{Password: password}
	resp, err := e.ValidateEndpoint(ctx, req)
	if err != nil {
		return false, err
	}
	valResp := resp.(validateResponse)
	if valResp.Err != nil {
		return false, valResp.Err
	}
	return valResp.Valid, nil
}

func (e Endpoints) HealthCheck() bool {
	return true
}

func NewVaultEndpoints(svc Service, logger logrus.Logger, trace stdopentracing.Tracer) Endpoints {
	fieldKeys := []string{"method"}
	svc = NewInstrumentingService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "ValeryPiashchynski",
			Subsystem: "vault_service",
			Name:      "request_count",
			Help:      "Number of requests received",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "ValeryPiashchynski",
			Subsystem: "vault_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds",
		}, fieldKeys),
		svc)

	var hashEndpoint endpoint.Endpoint
	{
		hashEndpoint = makeHashEndpoint(svc)
		hashEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(hashEndpoint)
		hashEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 10))(hashEndpoint)
		hashEndpoint = opentracing.TraceServer(trace, "hash")(hashEndpoint)
	}
	var validateEndpoint endpoint.Endpoint
	{
		validateEndpoint = makeValidateEndpoint(svc)
		validateEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(validateEndpoint)
		validateEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 10))(validateEndpoint)
		validateEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(validateEndpoint)
		validateEndpoint = opentracing.TraceServer(trace, "validate")(validateEndpoint)
	}
	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = makeHealthEndpoint(svc)
		healthEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(healthEndpoint)
		healthEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 10))(healthEndpoint)
		healthEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{Timeout: time.Duration(time.Second * 2)}))(healthEndpoint)
		healthEndpoint = opentracing.TraceServer(trace, "health")(healthEndpoint)
	}

	endpoints := Endpoints{
		HashEndpoint:        hashEndpoint,
		ValidateEndpoint:    validateEndpoint,
		HealthCheckEndpoint: healthEndpoint,
	}

	return endpoints
}
