package registration

import (
	"TaskManager/svcdiscovery"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	stdopentracing "github.com/opentracing/opentracing-go"
	kitprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	dbCreds      string
	consAddr     string
	vaultSvcName string
	svcTag       string
)

func init() {
	if dev := os.Getenv("DEV"); dev == "False" {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		dbCreds = viper.GetString("DbCreds.serverProd")
		consAddr = viper.GetString("Consul.addressProd")
		vaultSvcName = viper.GetString("services.vault") //to get hash from pass
		svcTag = viper.GetString("tags.tag")
	} else {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		dbCreds = viper.GetString("DbCreds.serverDev")
		consAddr = viper.GetString("Consul.addressDev")
		vaultSvcName = viper.GetString("services.vault") //to get hash from pass
		svcTag = viper.GetString("tags.tag")
	}

}

type Service interface {
	Registration(username, fullname, email, password string, isDisabled bool) (bool, error)
	UsernameValidation(username string) (bool, error)
	EmailValidation(email string) (bool, error)
	RegServiceHealthCheck() bool
}

func NewRegService() Service {
	return newRegService{}
}

type newRegService struct{}

type ServiceMiddleware func(svc Service) Service

//TODO marshall to structs
func (newRegService) Registration(username, fullname, email, password string, isDisabled bool) (ok bool, e error) {
	db, err := sql.Open("postgres", dbCreds)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var hresp hashResponse
	//TODO create struct
	var req []byte = []byte(`{"password":"` + password + `"}`)

	addr, err := svcdiscovery.ServiceDiscovery().Find(&consAddr, &vaultSvcName, &svcTag)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(addr+"/hash", "application/json; charset=utf-8", bytes.NewBuffer(req))
	if err != nil {
		return false, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &hresp)

	if hresp.Err != "" || err != nil {
		return false, err
	}

	_, err = db.Exec(`SET search_path = "xdev_site"`)
	if err != nil {
		return false, err.(*pq.Error)
	}

	stmIns, err := db.Query(`INSERT INTO "User" (username, fullname, email, passwordhash, isdisabled) VALUES ($1, $2, $3, $4, $5);`, username, fullname, email, hresp.Hash, false)
	if err != nil {
		return false, err.(*pq.Error)
	}
	defer stmIns.Close()

	return true, nil
}

func (newRegService) UsernameValidation(username string) (bool, error) {
	db, err := sql.Open("postgres", dbCreds)
	if err != nil {
		return false, err
	}

	defer db.Close()

	_, err = db.Exec(`SET search_path = "xdev_site"`)
	if err != nil {
		return false, err.(*pq.Error)
	}
	sel, err := db.Prepare(`SELECT id FROM "User" WHERE username = $1;`)
	if err != nil {
		panic(err.Error())
		return false, err.(*pq.Error)
	}
	defer sel.Close()

	var id int
	err = sel.QueryRow(username).Scan(&id)
	if err != nil { //NoRows error - user does no exist
		return false, nil
	} else {
		return true, nil // else - user exist
	}
}

func (newRegService) EmailValidation(email string) (bool, error) {
	db, err := sql.Open("postgres", dbCreds)
	if err != nil {
		return false, err
	}

	defer db.Close()

	_, err = db.Exec(`SET search_path = "xdev_site"`)
	if err != nil {
		return false, err.(*pq.Error)
	}

	sel, err := db.Prepare(`SELECT id FROM "User" WHERE email = $1;`)
	if err != nil {
		panic(err.Error())
		return false, err.(*pq.Error)
	}
	defer sel.Close()

	var id int

	err = sel.QueryRow(email).Scan(&id)
	if err != nil { //NoRows error - email does no exist
		return false, nil
	} else {
		return true, nil // else - email exist
	}

}

//TODO health create
func (newRegService) RegServiceHealthCheck() bool {
	return true
}

type Endpoints struct {
	RegEndpoint           endpoint.Endpoint
	UsernameValidEndpoint endpoint.Endpoint
	EmailValidEndpoint    endpoint.Endpoint
	RegHealthCheckEnpoint endpoint.Endpoint
}

func MakeRegEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RegRequest)

		ok, err := svc.Registration(req.Username, req.Fullname, req.Email, req.Password, req.isDisabled)
		if !ok {
			return nil, err
		}

		return RegResponce{Err: "", Status: ok}, nil
	}
}

func MakeUserValEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UsernameValidationRequest)

		exist, err := svc.UsernameValidation(req.User)
		if err != nil {
			return nil, err
		}
		return UsernameValidationResponce{Status: exist, Err: ""}, nil
	}
}

func MakeEmailValEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(EmailValidationRequest)

		exist, err := svc.EmailValidation(req.Email)
		if err != nil {
			return nil, err
		}

		return EmailValidationResponce{Status: exist, Err: ""}, nil
	}
}

func MakeRegHealthCheckEnpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(HealthRequest)

		v := svc.RegServiceHealthCheck()
		return HealthResponse{Status: v}, nil
	}
}

func NewEndpoints(svc Service, logger log.Logger, tracer stdopentracing.Tracer) Endpoints {

	fieldOps := []string{"method"}
	regSvcCounter := prometheus.NewCounterFrom(kitprometheus.CounterOpts{
		Namespace: "Adexin",
		Subsystem: "reg_service",
		Name:      "request_count",
		Help:      "Number of requests received",
	}, fieldOps)

	regSvcHist := prometheus.NewHistogramFrom(kitprometheus.HistogramOpts{
		Namespace: "Adexin",
		Subsystem: "reg_service",
		Name:      "request_latency_microseconds",
		Help:      "Number of requests received",
	}, fieldOps)

	svc = Metrics(regSvcCounter, regSvcHist)(svc)

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
		emailValidEndpoint = opentracing.TraceServer(tracer, "EmailValidation")(emailValidEndpoint)
		emailValidEndpoint = LoggingMiddleware(logger)(emailValidEndpoint)
	}

	var healthEnpoint endpoint.Endpoint
	{
		healthEnpoint = MakeRegHealthCheckEnpoint(svc)
		healthEnpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Limit(time.Second), 1))(healthEnpoint)
		healthEnpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(healthEnpoint)
		healthEnpoint = opentracing.TraceServer(tracer, "RegServiceHealthCheck")(healthEnpoint)
		healthEnpoint = LoggingMiddleware(logger)(healthEnpoint)
	}

	return Endpoints{
		RegEndpoint:           regEndpoint,
		UsernameValidEndpoint: usernameValidEndpoint,
		EmailValidEndpoint:    emailValidEndpoint,
		RegHealthCheckEnpoint: healthEnpoint,
	}
}
