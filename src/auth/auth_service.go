package auth

import (
	"context"
	"errors"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/kr/pretty"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"strings"
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

type ServiceMiddleware func(service Service) Service


var InvalidLoginErr = errors.New("username or password does not match, authentication failed")
var ErrRequestTypeNotFound = errors.New("request type only valid for login and logout")

//TODO check token in database
func (newService) Login(username, password string) (mesg string, roles []string, err error) {
	if strings.EqualFold("admin", username) &&
		strings.EqualFold("password", password) {
			mesg, roles, err = "Login succeed", []string{"Admin", "User"}, nil
	} else {
		mesg, roles, err = "", nil, InvalidLoginErr
	}
	return
}

//TODO remove token from db
func (newService) Logout() string {
	return "Logout Succeed"
}

func (newService) AuthHealtCheck() bool {
	return true
}

type CommonReqResp struct{

	TokenString string `json:"-"`
}

//request
type AuthRequest struct {
	CommonReqResp
	Username string `json:"username"`
	Password string `json:"password"`
	Type     string `json:"-"`
}

//response
type AuthResponse struct {
	CommonReqResp
	Roles []string `json:"roles,omitempty"`
	Mesg string `json:"mesg"`
	Err     error `json:"err,omitempty"`
}

//Health Request
type HealthRequest struct {

}

//Health Response
type HealthResponse struct {
	Status bool `json:"status"`
}

// endpoints wrapper
type Endpoints struct {
	AuthEndpoint endpoint.Endpoint
	HealthEndpoint endpoint.Endpoint
}



func MakeLoginEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var (
			roles []string
			mesg string
			err error
		)
		req := request.(AuthRequest)
		pretty.Print("ctx")
		if strings.EqualFold(req.Type, "login") {
			mesg, roles, err = svc.Login(req.Username, req.Password)
		} else if strings.EqualFold(req.Type, "logout") {
			mesg = svc.Logout()
			err = nil
		} else {
			return nil, ErrRequestTypeNotFound
		}

		// check if err is not null
		if err != nil {
			return nil, err
		}
		return AuthResponse{Mesg:mesg, Roles: roles, Err: err}, nil
	}
}

// creating health endpoint
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
		loginEndpoint = JwtEndpoint("localhost","8500",logger)(loginEndpoint)
		loginEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(loginEndpoint)
		loginEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(loginEndpoint)
		loginEndpoint = opentracing.TraceServer(trace, "login")(loginEndpoint)
		loginEndpoint = LoggingMiddleware(log.With(logger, "method","login"))(loginEndpoint)
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
		AuthEndpoint: loginEndpoint,
		HealthEndpoint: healthEndpoint,
	}
}
































