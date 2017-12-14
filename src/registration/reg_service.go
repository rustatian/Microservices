package registration

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

var dbCreds string

func init() {
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	dbCreds = viper.GetString("DbCreds.server")
}

type Service interface {
	Registration(username, fullname, email, passwordHash, jwtToken string, isDisabled bool) (bool, error)
	UsernameValidation()
	EmailValidation()
}

func NewRegService() Service {
	return newRegService{}
}

type newRegService struct{}

func (newRegService) Registration(username, fullname, email, passwordHash, jwtToken string, isDisabled bool) (bool, error) {
	db, err := sql.Open("mysql", dbCreds)
	if err != nil {
		return false, err
	}

	defer db.Close()

	stmIns, err := db.Prepare("INSERT INTO USER (Username, FullName, email, PasswordHash, jwtToken, IsDisabled) VALUES (?, ?, ?, ?, ?, ?);")
	defer stmIns.Close()

	_, err = stmIns.Exec(username, fullname, email, passwordHash, jwtToken, isDisabled)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (newRegService) UsernameValidation() {

}

func (newRegService) EmailValidation() {

}

type RegRequest struct {
	Username     string `json:"username"`
	Fullname     string `json:"fullname"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	JwtToken     string `json:"jwt_token"`
	isDisabled   bool   `json:"is_disabled"`
}

type RegResponce struct {
	Status bool   `json:"status"`
	Err    string `json:"err, omitempty"`
}

type UsernameValidationRequest struct {
	User string `json:"user"`
}

type UsernameValidationResponce struct {
	Status string `json:"status"`
	Err    string `json:"err, omitempty"`
}

type EmailValidationRequest struct {
	Email string `json:"email"`
}

type EmailValidationResponce struct {
	Status string `json:"status"`
	Err    string `json:"err, omitempty"`
}

type Endpoints struct {
	RegEndpoint           endpoint.Endpoint
	UsernameValidEndpoint endpoint.Endpoint
	EmailValidEndpoint    endpoint.Endpoint
}

func MakeRegEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, err := request.(RegRequest)
		if err != nil {
			return nil, err
		}

		ok, err := svc.Registration(req.Username, req.Fullname, req.Email, req.PasswordHash, req.JwtToken, req.isDisabled)
		if !ok {
			return nil, err
		}

		return RegResponce{Err: "", Status: ok}, nil
	}
}

//TODO implement
func MakeUserValEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return "", nil
	}
}


func MakeEmailValEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return "", nil
	}
}

func NewEnpoints(svc Service, logger log.Logger, tracer stdopentracing.Tracer) Endpoints {
	var regEndpoint endpoint.Endpoint
	{
		regEndpoint = MakeRegEndpoint(svc)
		regEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(regEndpoint)
		regEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(regEndpoint)
		regEndpoint = opentracing.TraceServer(tracer, "Registration")(regEndpoint)
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
		emailValidEndpoint = opentracing.TraceServer(tracer,"EmailValidation")(emailValidEndpoint)
		emailValidEndpoint = LoggingMiddleware(logger)(emailValidEndpoint)
	}

	return Endpoints {
		RegEndpoint: regEndpoint,
		UsernameValidEndpoint: usernameValidEndpoint,
		EmailValidEndpoint: emailValidEndpoint,
	}
}
