package main

import (
	"TaskManager/src/vault"
	"context"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		consulAddr = flag.String("consul.addr", "localhost", "consul address")
		consulPort = flag.String("consul.port", ":8500", "consul port")
		vaultAddr = flag.String("vault.addr", "localhost", "advertise address")
		vaultPort = flag.String("vault.port", ":10000", "advertise port")
		)
	flag.Parse()
	ctx := context.Background()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	svc := vault.NewVaultService()
	//Just try different way
	//svc = vault.LoggingMiddleware(logger)(svc)

	tracer := stdopentracing.GlobalTracer()
	reg := vault.Register(*consulAddr, *consulPort, *vaultAddr,*vaultPort, "vaultsvc", logger)


	endpoints := vault.NewEndpoints(svc, logger, tracer)
	r := vault.MakeVaultHttpHandler(ctx, endpoints, logger)


	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()


	// HTTP transport.
	go func() {
		reg.Register()
		logger.Log("transport", "HTTP", "addr", *vaultPort)
		errc <- http.ListenAndServe(*vaultPort, r)
	}()


	logger.Log("exit", <-errc)
	reg.Deregister()
}
