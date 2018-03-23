package registration

import (
	"context"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	_ "github.com/lib/pq"
	stdopentracing "github.com/opentracing/opentracing-go"
	kitprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type Endpoints struct {
	RegEndpoint            endpoint.Endpoint
	UsernameValidEndpoint  endpoint.Endpoint
	EmailValidEndpoint     endpoint.Endpoint
	RegHealthCheckEndpoint endpoint.Endpoint
}

func MakeRegEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RegRequest)

		ok, err := svc.Registration(ctx, req.Username, req.Fullname, req.Email, req.Password, req.IsDisabled)
		if !ok {
			return nil, err
		}

		return RegResponce{Err: "", Status: ok}, nil
	}
}

func MakeUserValEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UsernameValidationRequest)

		exist, err := svc.UsernameValidation(ctx, req.User)
		if err != nil {
			return nil, err
		}
		return UsernameValidationResponce{Status: exist, Err: ""}, nil
	}
}

func MakeEmailValEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(EmailValidationRequest)

		exist, err := svc.EmailValidation(ctx, req.Email)
		if err != nil {
			return nil, err
		}

		return EmailValidationResponce{Status: exist, Err: ""}, nil
	}
}

func MakeRegHealthCheckEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(HealthRequest)

		v := svc.RegServiceHealthCheck()
		return HealthResponse{Status: v}, nil
	}
}

func NewEndpoints(svc Service, logger log.Logger, tracer stdopentracing.Tracer) Endpoints {
	fieldKeys := []string{"method"}
	regSvcCounter := prometheus.NewCounterFrom(kitprometheus.CounterOpts{
		Namespace: "ValeryPiashchynski",
		Subsystem: "reg_service",
		Name:      "request_count",
		Help:      "Number of requests received",
	}, fieldKeys)

	regSvcHist := prometheus.NewHistogramFrom(kitprometheus.HistogramOpts{
		Namespace: "ValeryPiashchynski",
		Subsystem: "reg_service",
		Name:      "request_latency_microseconds",
		Help:      "Number of requests received",
	}, fieldKeys)

	svc = NewInstrumentingService(regSvcCounter, regSvcHist, svc)

	var regEndpoint endpoint.Endpoint
	{
		regEndpoint = MakeRegEndpoint(svc)
		regEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(regEndpoint)
		regEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(regEndpoint)
		regEndpoint = opentracing.TraceServer(tracer, "RegistrationViaHTTP")(regEndpoint)
		regEndpoint = LoggingMiddleware(logger)(regEndpoint)
	}

	var usernameValidEndpoint endpoint.Endpoint
	{
		usernameValidEndpoint = MakeUserValEndpoint(svc)
		usernameValidEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(usernameValidEndpoint)
		usernameValidEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(usernameValidEndpoint)
		usernameValidEndpoint = opentracing.TraceServer(tracer, "UsernameValidation")(usernameValidEndpoint)
		usernameValidEndpoint = LoggingMiddleware(logger)(usernameValidEndpoint)
	}

	var emailValidEndpoint endpoint.Endpoint
	{
		emailValidEndpoint = MakeEmailValEndpoint(svc)
		emailValidEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(emailValidEndpoint)
		emailValidEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(emailValidEndpoint)
		emailValidEndpoint = opentracing.TraceServer(tracer, "EmailValidation")(emailValidEndpoint)
		emailValidEndpoint = LoggingMiddleware(logger)(emailValidEndpoint)
	}

	var healthEnpoint endpoint.Endpoint
	{
		healthEnpoint = MakeRegHealthCheckEndpoint(svc)
		healthEnpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Limit(time.Second), 1))(healthEnpoint)
		healthEnpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(healthEnpoint)
		healthEnpoint = opentracing.TraceServer(tracer, "RegServiceHealthCheck")(healthEnpoint)
		healthEnpoint = LoggingMiddleware(logger)(healthEnpoint)
	}

	return Endpoints{
		RegEndpoint:            regEndpoint,
		UsernameValidEndpoint:  usernameValidEndpoint,
		EmailValidEndpoint:     emailValidEndpoint,
		RegHealthCheckEndpoint: healthEnpoint,
	}
}
