package calendar

import (
	"context"
	"database/sql"
	"encoding/json"
	stdjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/lib/pq"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"os"
	"time"
)

var (
	dbCreds string
	secKey  string
)

var (
	signingMethod = stdjwt.SigningMethodHS256
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
		secKey = viper.GetString("SecretKey.key")

	} else {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		dbCreds = viper.GetString("DbCreds.serverDev")
		secKey = viper.GetString("SecretKey.key")
	}
}

type TaskService interface {
	GetTasks(ctx context.Context, username string, tr timeRange) (resp string, e error)
	HealthChecks() bool
}

type ServiceMiddleware func(svc TaskService) TaskService

type tCalendarService struct{}

func NewService() TaskService {
	return tCalendarService{}
}

func (s tCalendarService) GetTasks(ctx context.Context, username string, tr timeRange) (resp string, e error) {
	switch tr {
	case TDay:

	case TWeek:

	case TMonth:
		db, err := sql.Open("postgres", dbCreds)
		defer db.Close()
		if err != nil {
			err = err.(*pq.Error)
			return "", err
		}

		//Get current time
		currTime := time.Now()
		year, month, day := currTime.Date()

		//Get prev month and next month
		prevMnt := time.Date(year, month-1, day, 0, 0, 0, 0, time.UTC).Unix()
		futureMnt := time.Date(year, month+1, day, 0, 0, 0, 0, time.UTC).Unix()

		_, err = db.Exec(`SET search_path = "xdev_site"`)
		if err != nil {
			return "", err.(*pq.Error)
		}
		sel, err := db.Query(`SELECT taskid, taskcaption, taskdescription, tfrom, tto FROM calendartasks WHERE userid = (SELECT id FROM "User" WHERE username = $1) AND tfrom >= $2 AND tto <= $3`, username, prevMnt, futureMnt)
		if err != nil {
			return "", err
		}
		defer sel.Close()

		var tasks []Task

		for sel.Next() {
			var taskId int
			var taskCaption string
			var taskDescription string
			var from uint64
			var to uint64

			err = sel.Scan(&taskId, &taskCaption, &taskDescription, &from, &to)
			if err != nil {
				return "", err.(*pq.Error)
			}

			tasks = append(tasks, Task{
				TaskId:          taskId,
				TaskCaption:     taskCaption,
				TaskDescription: taskDescription,
				From:            from,
				To:              to,
			})
		}

		data, err := json.Marshal(&tasks)
		if err != nil {
			return "", err
		}

		return string(data), nil

	case TYear:

	default:

	}

	return "", nil
}

func (s tCalendarService) HealthChecks() (hc bool) {
	return true
}

type Endpoints struct {
	TaskCalendarEnpoint endpoint.Endpoint
	HealthChecks        endpoint.Endpoint
}

func (e Endpoints) GetTasks(ctx context.Context, username string, tr timeRange) (resp string, er error) {
	req := tasksRequest{
		TimeRange: int(tr),
		User:      username,
	}

	rsp, err := e.TaskCalendarEnpoint(ctx, req)
	if err != nil {
		return "", err
	}

	tcresp := rsp.(tasksResponce)
	if tcresp.Err != "" {
		return "", err
	}
	return "", nil
}

func MakeHealthEndpoint(svc TaskService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		v := svc.HealthChecks()
		return healthResponse{Status: v}, nil
	}
}

func MakeTasksEndpoint(svc TaskService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(tasksRequest)
		v, err := svc.GetTasks(ctx, req.User, timeRange(req.TimeRange))
		if err != nil {
			return tasksResponce{Err: err.Error()}, err
		}
		return tasksResponce{Tasks: v, Err: ""}, nil
	}
}

func NewEndpoints(svc TaskService, logger log.Logger, trace stdopentracing.Tracer) Endpoints {
	fieldKeys := []string{"method"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "ValeryPiashchynski",
		Subsystem: "vault_service",
		Name:      "request_count",
		Help:      "Number of requests received",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "ValeryPiashchynski",
		Subsystem: "vault_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds",
	}, fieldKeys)

	svc = Metrics(requestCount, requestLatency)(svc)

	var getTasksEnpoint endpoint.Endpoint
	{
		getTasksEnpoint = MakeTasksEndpoint(svc)
		getTasksEnpoint = jwt.NewParser(keyFunc(), signingMethod, jwt.MapClaimsFactory)(getTasksEnpoint)
		getTasksEnpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(getTasksEnpoint)
		getTasksEnpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{Timeout: time.Duration(time.Second * 2)}))(getTasksEnpoint)
		getTasksEnpoint = opentracing.TraceServer(trace, "GetTasks")(getTasksEnpoint)
		getTasksEnpoint = LoggingMiddleware(log.With(logger, "method", "GetTasks"))(getTasksEnpoint)
	}

	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = MakeHealthEndpoint(svc)
		healthEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(healthEndpoint)
		healthEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{Timeout: time.Duration(time.Second * 2)}))(healthEndpoint)
		healthEndpoint = opentracing.TraceServer(trace, "Health")(healthEndpoint)
		healthEndpoint = LoggingMiddleware(log.With(logger, "method", "health"))(healthEndpoint)
	}

	return Endpoints{
		getTasksEnpoint,
		healthEndpoint,
	}
}

func keyFunc() stdjwt.Keyfunc {
	return func(token *stdjwt.Token) (interface{}, error) {
		return []byte(secKey), nil
	}
}
