package authorization

import (
	"TaskManager/svcdiscovery"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	_ "github.com/go-sql-driver/mysql"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
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
	Login(username, password string) (mesg string, roles []string, err error)
	Logout() string
	AuthHealtCheck() bool
}

func NewAuthService() Service {
	return newService{}
}

type ServiceMiddleware func(svc Service) Service

type newService struct{}

func (newService) Login(username, password string) (mesg string, roles []string, err error) {
	db, err := sql.Open("mysql", dbCreds)
	if err != nil {
		return "", nil, err
	}
	defer db.Close()

	if err != nil {
		return "", nil, err
	}

	sel, err := db.Prepare("SELECT ID FROM User WHERE Username = ?;")
	if err != nil {
		return "", nil, err
	}
	defer sel.Close()

	var resId int

	err = sel.QueryRow(username).Scan(&resId)
	if err != nil {
		return "Login Failed", nil, fmt.Errorf("user does't exist")
	} else {

		pass, err := db.Prepare("SELECT PasswordHash FROM User WHERE ID = ?")
		defer pass.Close()

		var hash string
		pass.QueryRow(resId).Scan(&hash)

		var vresp validateResponse

		var req []byte = []byte(`{"password":"` + password + `", "hash": "` + hash + `"}`)
		addr, err := svcdiscovery.ServiceDiscovery().Find(&consAddr, &vaultSvcName, &svcTag)
		if err != nil {
			return "", nil, err
		}

		resp, err := http.Post(addr+"/validate", "application/json; charset=utf-8", bytes.NewBuffer(req))
		if err != nil {
			return "Login request failed, password error", nil, err
		}

		body, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &vresp)

		if vresp.Err != "" || vresp.Valid == false {
			return "Login failed, password validate error", nil, fmt.Errorf("password validation error")
		}
		if err != nil {
			return "Login failed, password validate error", nil, err
		}

		return "Login succeed, password ok", []string{"Admin", "User"}, nil
	}
}

func (newService) Logout() string {
	return "Logout Succeed"
}

//TODO create full check
func (newService) AuthHealtCheck() bool {
	return true
}

type Endpoints struct {
	LoginEndpoint  endpoint.Endpoint
	LogoutEnpoint  endpoint.Endpoint
	HealthEndpoint endpoint.Endpoint
}

func MakeLoginEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var (
			roles []string
			mesg  string
			err   error
		)

		req := request.(LoginRequest)
		mesg, roles, err = svc.Login(req.Username, req.Password)

		if err != nil {
			return nil, err
		}
		return LoginResponce{Mesg: mesg, Roles: roles, Err: ""}, nil
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
		return HealthResponse{Status: status}, nil
	}
}

func NewEndpoints(svc Service, logger log.Logger, trace stdopentracing.Tracer) Endpoints {

	fieldKeys := []string{"method"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "Adexin",
		Subsystem: "auth_service",
		Name:      "request_count",
		Help:      "Number of requests received",
	}, fieldKeys)

	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "Adexin",
		Subsystem: "auth_service",
		Name:      "request_latency",
		Help:      "Total duration of requests in microseconds",
	}, fieldKeys)

	svc = Metrics(requestCount, requestLatency)(svc)

	var loginEndpoint endpoint.Endpoint
	{
		loginEndpoint = MakeLoginEndpoint(svc)
		loginEndpoint = JwtLoginEndpoint(logger)(loginEndpoint)
		loginEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(loginEndpoint)
		loginEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(loginEndpoint)
		loginEndpoint = opentracing.TraceServer(trace, "login")(loginEndpoint)
		loginEndpoint = LoggingMiddleware(log.With(logger, "method", "login"))(loginEndpoint)
	}

	var logoutEndpoint endpoint.Endpoint
	{
		logoutEndpoint = MakeLogoutEndpoint(svc)
		logoutEndpoint = JwtLogoutEndpoint(logger)(logoutEndpoint)
		logoutEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(logoutEndpoint)
		logoutEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(logoutEndpoint)
		logoutEndpoint = opentracing.TraceServer(trace, "logout")(logoutEndpoint)
		logoutEndpoint = LoggingMiddleware(log.With(logger, "method", "logout"))(logoutEndpoint)
	}

	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = MakeHealthEndpoint(svc)
		healthEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(healthEndpoint)
		healthEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(healthEndpoint)
		healthEndpoint = opentracing.TraceServer(trace, "health")(healthEndpoint)
		healthEndpoint = LoggingMiddleware(log.With(logger, "method", "health"))(healthEndpoint)
	}

	return Endpoints{
		LoginEndpoint:  loginEndpoint,
		LogoutEnpoint:  logoutEndpoint,
		HealthEndpoint: healthEndpoint,
	}
}
