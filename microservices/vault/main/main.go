package main

import (
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ValeryPiashchynski/TaskManager/microservices/tools"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/ratelimit"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/ValeryPiashchynski/TaskManager/microservices/vault"
	"github.com/ValeryPiashchynski/TaskManager/svcdiscovery"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

func main() {
	var (
		consulAddr = flag.String("consul.addr", "localhost", "consul address")
		consulPort = flag.String("consul.port", "8500", "consul port")
		vaultPort  = flag.String("vault.port", "10000", "vault port")
		svcName    = flag.String("service.name", "vaultsvc", "Vault service name")
	)

	vaultAddr, _ := externalIP()

	flag.Parse()

	logg := logrus.New()
	logg.Out = os.Stdout

	ctx := &vault.Context{
		Log: logg,
	}

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	pwdChecker := tools.NewPasswordChecker()
	svc := vault.NewVaultService(pwdChecker)

	tracer := stdopentracing.GlobalTracer()
	reg := svcdiscovery.ServiceDiscovery().RegistrationViaHTTP(*consulAddr, *consulPort, vaultAddr, *vaultPort, *svcName, logger)
	defer reg.Deregister()

	endpoints := NewService(svc, logger, tracer)

	errCh := make(chan error)
	// Interrupt handler.
	c := make(chan os.Signal)

	//Error handler
	go func() {
		//logger.Log("nats error:", <-errCh)
		stdlog.Fatal(<-errCh)
	}()

	r := vault.MakeVaultHttpHandler(endpoints, logger)
	srv := &http.Server{
		Handler:      vault.NewContextHandler(ctx, r),
		Addr:         vaultAddr + ":" + *vaultPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errCh <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		reg.Register()
		logger.Log("transport", "HTTP", "addr", ":"+*vaultPort)

		//Custom server with logrus
		errCh <- srv.ListenAndServe()
	}()

	logger.Log("exit", <-errCh)
}

func NewService(svc vault.Service, logger log.Logger, trace stdopentracing.Tracer) vault.Endpoints {
	//declare metrics
	fieldKeys := []string{"method"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "Adexin",
		Subsystem: "vault_service",
		Name:      "request_count",
		Help:      "Number of requests received",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "Adexin",
		Subsystem: "vault_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds",
	}, fieldKeys)

	svc = vault.Metrics(requestCount, requestLatency)(svc)
	svc = vault.NewLoggingService(log.With(logger, "hash", "validate", "health"), svc)

	var hashEndpoint endpoint.Endpoint
	{
		hashEndpoint = vault.MakeHashEndpoint(svc)
		hashEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 10))(hashEndpoint)
		hashEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(hashEndpoint)
		hashEndpoint = opentracing.TraceServer(trace, "hash")(hashEndpoint)
	}
	var validateEndpoint endpoint.Endpoint
	{
		validateEndpoint = vault.MakeValidateEndpoint(svc)
		validateEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 10))(validateEndpoint)
		validateEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(validateEndpoint)
		validateEndpoint = opentracing.TraceServer(trace, "validate")(validateEndpoint)
	}
	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = vault.MakeHealthEndpoint(svc)
		healthEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Millisecond), 10))(healthEndpoint)
		healthEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{Timeout: time.Duration(time.Second * 2)}))(healthEndpoint)
		healthEndpoint = opentracing.TraceServer(trace, "health")(healthEndpoint)
	}

	return vault.Endpoints{
		ValidateEndpoint:    validateEndpoint,
		VaultHealthEndpoint: healthEndpoint,
		HashEndpoint:        hashEndpoint,
	}
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("not connected")
}
