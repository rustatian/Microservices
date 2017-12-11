package main

import (
	"TaskManager/src/auth"
	"context"
	"flag"
	"fmt"
	stdopentracing "github.com/opentracing/opentracing-go"
	ilog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
)

func main() {
	// parse variable from input command
	var (
		consulAddr = flag.String("consul.addr", "localhost", "consul address")
		consulPort = flag.String("consul.port", ":8500", "consul port")
		authAddr   = flag.String("advertise.addr", "localhost", "advertise address")
		authPort   = flag.String("advertise.port", ":10001", "advertise port")
	)
	flag.Parse()
	ctx := context.Background()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	//var svc auth.Service
	svc := auth.NewAuthService()
	tracer := stdopentracing.GlobalTracer()

	//svc = auth.LoggingMiddleware(logger)(svc)

	endpoints := auth.NewEndpoints(svc, logger, tracer)
	//endpoints2 := auth.JwtEndpoint(*consulAddr, *consulPort, logger)

	endpoint := auth.Endpoints{
		AuthEndpoint:   endpoints.AuthEndpoint,
		HealthEndpoint: endpoints.HealthEndpoint,
	}

	r := auth.MakeAuthHttpHandler(ctx, endpoint, logger)

	// Register Service to Consul
	reg := auth.Register(*consulAddr, *consulPort, *authAddr, *authPort, "authsvc", logger)

	errChan := make(chan error)
	// HTTP transport
	go func() {
		ilog.Println("Starting server at port", *authPort)
		// register service
		reg.Register()
		handler := r
		errChan <- http.ListenAndServe(*authPort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	chErr := <-errChan

	reg.Deregister()
	ilog.Fatalln(chErr)
}
