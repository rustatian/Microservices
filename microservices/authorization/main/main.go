package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/ValeryPiashchynski/TaskManager/microservices/authorization"
	"github.com/ValeryPiashchynski/TaskManager/svcdiscovery"
	"github.com/go-kit/kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
	ilog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		consulAddr = flag.String("consul.addr", "127.0.0.1", "consul address")
		consulPort = flag.String("consul.port", "8500", "consul port")
		//authAddr   = flag.String("auth.addr", "localhost", "auth address")
		authPort = flag.String("auth.port", "10001", "auth port")
		svcName  = flag.String("service.name", "authsvc", "Authorization microservice name")
	)

	authAddr, _ := externalIP()

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
	reg := svcdiscovery.ServiceDiscovery().Registration(*consulAddr, *consulPort, authAddr, *authPort, *svcName, logger)

	errChan := make(chan error)
	defer close(errChan)

	// HTTP transport
	go func() {
		ilog.Println("Starting server at port", *authPort)
		reg.Register()
		handler := r
		errChan <- http.ListenAndServe(net.JoinHostPort(authAddr, *authPort), handler)
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
