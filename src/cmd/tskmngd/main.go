package main

import (
	"flag"
	"TaskManager/src/vault"
	consulsd "github.com/go-kit/kit/sd/consul"
	stdopentracing "github.com/opentracing/opentracing-go"
	"os"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
	"context"
	"os/signal"
	"syscall"
	"net/http"
)

func main() {
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		consulAddr   = flag.String("consul.addr", "", "Consul agent address")
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

	var service = vault.NewVaultService()
	var tracer = stdopentracing.GlobalTracer()
	register := vault.Register("localhost", *httpAddr,"vaultsvc", client, logger)

	ctx := context.Background()
	//r := mux.NewRouter()


	endpoints := vault.NewEndpoints(service, logger, tracer)
	r := vault.MakeHttpHandler(ctx, endpoints, logger)


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
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, r)
	}()


	logger.Log("exit", <-errc)
	register.Deregister()
}

