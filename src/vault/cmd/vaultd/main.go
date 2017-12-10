package main

import (
	"TaskManager/src/vault"
	"context"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	stdopentracing "github.com/opentracing/opentracing-go"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		vaultHttpPort = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		consulAddr    = flag.String("consul.addr", "", "Consul agent address")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		if len(*consulAddr) > 0 {
			consulConfig.Address = *consulAddr
		}
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	service := vault.NewVaultService()
	//Just try different way
	service = vault.LoggingMiddleware(logger)(service)

	tracer := stdopentracing.GlobalTracer()
	register := vault.Register("localhost", *vaultHttpPort,"vaultsvc", client, logger)

	ctx := context.Background()

	endpoints := vault.NewEndpoints(service, logger, tracer)
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
		register.Register()
		logger.Log("transport", "HTTP", "addr", *vaultHttpPort)
		errc <- http.ListenAndServe(*vaultHttpPort, r)
	}()


	logger.Log("exit", <-errc)
	register.Deregister()
}
