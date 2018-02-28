package vault

import (
	"context"

	"github.com/go-kit/kit/endpoint"
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

//func NewService(svc vault.Service, logger log.Logger, trace stdopentracing.Tracer) {
//	//declare metrics
//	svc = vault.NewLoggingService(log.With(logger, "hash", "validate", "health"), svc)
//
//	var hashEndpoint endpoint.Endpoint
//	{
//		hashEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 10))(hashEndpoint)
//		hashEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(hashEndpoint)
//		hashEndpoint = opentracing.TraceServer(trace, "hash")(hashEndpoint)
//	}
//	var validateEndpoint endpoint.Endpoint
//	{
//		validateEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 10))(validateEndpoint)
//		validateEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(validateEndpoint)
//		validateEndpoint = opentracing.TraceServer(trace, "validate")(validateEndpoint)
//	}
//	var healthEndpoint endpoint.Endpoint
//	{
//		healthEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 10))(healthEndpoint)
//		healthEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{Timeout: time.Duration(time.Second * 2)}))(healthEndpoint)
//		healthEndpoint = opentracing.TraceServer(trace, "health")(healthEndpoint)
//	}
//
//}
