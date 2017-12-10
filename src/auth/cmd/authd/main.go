package main

import (
	"TaskManager/src/auth"
	"context"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	ilog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// parse variable from input command
	var (
		consulAddr = flag.String("consul.addr", "localhost", "consul address")
		consulPort = flag.String("consul.port", "8500", "consul port")
		advertiseAddr = flag.String("advertise.addr", "localhost", "advertise address")
		advertisePort = flag.String("advertise.port", "3000", "advertise port")
	)
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	//var svc auth.Service
	svc := auth.NewAuthService()
	svc = auth.LoggingMiddleware(logger)(svc)

	e := auth.MakeLoginEndpoint(svc)
	e = auth.JwtEndpoint(*consulAddr, *consulPort, logger)(e)

	endpoint := auth.Endpoints{
		AuthEndpoint: e,
		HealthEndpoint: auth.MakeHealthEndpoint(svc),
	}

	r := auth.MakeHttpHandler(ctx, endpoint, logger)

	// Register Service to Consul
	registar := auth.Register(*consulAddr,
		*consulPort,
		*advertiseAddr,
		*advertisePort)

	// HTTP transport
	go func() {
		ilog.Println("Starting server at port", *advertisePort)
		// register service
		registar.Register()
		handler := r
		errChan <- http.ListenAndServe( ":" + *advertisePort, handler)
	}()


	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	chErr := <- errChan


	registar.Deregister()
	ilog.Fatalln(chErr)
}
