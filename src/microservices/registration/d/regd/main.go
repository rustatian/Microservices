package main

import (
	"TaskManager/src/registration"
	"context"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	ilog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"TaskManager/src/svcdiscovery"
	"net"
)

func main() {
	var (
		consulAddr = flag.String("consul.addr", "localhost", "consul address")
		consulPort = flag.String("consul.port", "8500", "consul port")
		regAddr    = flag.String("reg.addr", "localhost", "reg address")
		regPort    = flag.String("reg.port", "10002", "reg port")
		svcName    = flag.String("service.name", "regsvc", "Registration microservice name")
	)

	flag.Parse()
	ctx := context.Background()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	svc := registration.NewRegService()
	tracer := opentracing.GlobalTracer()

	endpoints := registration.NewEnpoints(svc, logger, tracer)

	endpoint := registration.Endpoints{
		RegEndpoint:           endpoints.RegEndpoint,
		UsernameValidEndpoint: endpoints.UsernameValidEndpoint,
		EmailValidEndpoint:    endpoints.EmailValidEndpoint,
		RegHealthCheckEnpoint: endpoints.RegHealthCheckEnpoint,
	}

	r := registration.MakeRegHttpHandler(ctx, endpoint, logger)
	reg := svcdiscovery.ServiceDiscovery().Registration(*consulAddr, *consulPort, *regAddr, *regPort, *svcName, logger)

	errChan := make(chan error)
	defer close(errChan)

	go func() {
		ilog.Println("Starting server at port", *regPort)
		reg.Register()
		handler := r
		errChan <- http.ListenAndServe(net.JoinHostPort(*regAddr, *regPort), handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	var e error = <-errChan
	reg.Deregister()
	ilog.Fatalln(e)
}
