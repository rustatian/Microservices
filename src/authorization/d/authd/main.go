package main

import (
	"TaskManager/src/authorization"
	"context"
	"flag"
	"fmt"
	ilog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
)

func main() {
	var (
		consulAddr = flag.String("consul.addr", "localhost", "consul address")
		consulPort = flag.String("consul.port", ":8500", "consul port")
		authAddr   = flag.String("auth.addr", "localhost", "auth address")
		authPort   = flag.String("auth.port", ":10001", "auth port")
	)

	flag.Parse()
	ctx := context.Background()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	svc := authorization.NewAuthService()
	tracer := stdopentracing.GlobalTracer()

	endpoints := authorization.NewEndpoints(svc, logger, tracer)

	endpoint := authorization.Endpoints{
		LoginEndpoint:  endpoints.LoginEndpoint,
		LogoutEnpoint:  endpoints.LogoutEnpoint,
		HealthEndpoint: endpoints.HealthEndpoint,
	}

	r := authorization.MakeAuthHttpHandler(ctx, endpoint, logger)

	// Register Service to Consul
	reg := authorization.Register(*consulAddr, *consulPort, *authAddr, *authPort, logger)

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
