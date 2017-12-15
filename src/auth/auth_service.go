package auth

import (
	"context"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"time"
)

type Service interface {
	Login(username, password string) (mesg string, roles []string, err error)
	Logout() string
	AuthHealtCheck() bool
}

func NewAuthService() Service {
	return newService{}
}



type newService struct {}

//TODO check username and pass in database
func (newService) Login(username, password string) (mesg string, roles []string, err error) {
	mesg, roles, err = "Login succeed", []string{"Admin", "User"}, nil
	//mesg, roles, err = "", nil, InvalidLoginErr
	return
}

//TODO remove token from db
func (newService) Logout() string {
	return "Logout Succeed"
}

//TODO check username
func (newService) ValidateUsername() bool {
  return true
}

func (newService) AuthHealtCheck() bool {
	return true
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TokenString string `json:"token_string"`
}

type LoginResponce struct {
	Roles []string `json:"roles, omitempty"`
	Mesg string `json:"mesg"`
	TokenString string `json:"token_string"`
	Err string `json:"err, omitempty"`
}

type LogoutRequest struct {
	TokenString string `json:"token_string"`
	Username string `json:"username"`
}

type LogoutResponce struct {
	Status bool `json:"status"`
}

type HealthRequest struct {

}

type HealthResponse struct {
	Status bool `json:"status"`
}

// endpoints wrapper
type Endpoints struct {
	LoginEndpoint endpoint.Endpoint
	LogoutEnpoint endpoint.Endpoint
	HealthEndpoint endpoint.Endpoint
}


func MakeLoginEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var (
			roles []string
			mesg string
			err error
		)

		req := request.(LoginRequest)
		mesg, roles, err = svc.Login(req.Username, req.Password)

		if err != nil {
			return nil, err
		}
		return LoginResponce { Mesg:mesg, Roles: roles, Err: "" }, nil
	}
}

func MakeLogoutEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		//req := request.(AuthRequest)
		svc.Logout()
		return "", nil
	}
}

func MakeHealthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		status := svc.AuthHealtCheck()
		return HealthResponse{Status: status }, nil
	}
}

func NewEndpoints(svc Service, logger log.Logger, trace stdopentracing.Tracer) Endpoints {
	var loginEndpoint endpoint.Endpoint
	{
		loginEndpoint = MakeLoginEndpoint(svc)
		loginEndpoint = JwtLoginEndpoint(logger)(loginEndpoint)
		loginEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(loginEndpoint)
		loginEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(loginEndpoint)
		loginEndpoint = opentracing.TraceServer(trace, "login")(loginEndpoint)
		loginEndpoint = LoggingMiddleware(log.With(logger, "method","login"))(loginEndpoint)
	}

	var logoutEndpoint endpoint.Endpoint
	{
		logoutEndpoint = MakeLogoutEndpoint(svc)
		logoutEndpoint = JwtLogoutEndpoint(logger)(logoutEndpoint)
		logoutEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(logoutEndpoint)
		logoutEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(logoutEndpoint)
		logoutEndpoint = opentracing.TraceServer(trace, "logout")(logoutEndpoint)
		logoutEndpoint = LoggingMiddleware(log.With(logger, "method","logout"))(logoutEndpoint)
	}

	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = MakeHealthEndpoint(svc)
		healthEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(healthEndpoint)
		healthEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(healthEndpoint)
		healthEndpoint = opentracing.TraceServer(trace, "health")(healthEndpoint)
		healthEndpoint = LoggingMiddleware(log.With(logger, "method","health"))(healthEndpoint)

	}

	return Endpoints{
		LoginEndpoint: loginEndpoint,
		LogoutEnpoint: logoutEndpoint,
		HealthEndpoint: healthEndpoint,
	}
}
































