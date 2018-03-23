package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/ValeryPiashchynski/TaskManager/microservices/registration"
	"github.com/ValeryPiashchynski/TaskManager/svcdiscovery"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	ilog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		consulAddr = flag.String("consul.addr", "localhost", "consul address")
		consulPort = flag.String("consul.port", "8500", "consul port")
		regPort    = flag.String("reg.port", "10002", "reg port")
		svcName    = flag.String("service.name", "regsvc", "RegistrationViaHTTP microservice name")
	)

	flag.Parse()
	regAddr, _ := externalIP()

	ctx := context.Background()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	svc := registration.NewRegService()
	tracer := opentracing.GlobalTracer()

	endpoints := registration.NewEndpoints(svc, logger, tracer)

	endpoint := registration.Endpoints{
		RegEndpoint:            endpoints.RegEndpoint,
		UsernameValidEndpoint:  endpoints.UsernameValidEndpoint,
		EmailValidEndpoint:     endpoints.EmailValidEndpoint,
		RegHealthCheckEndpoint: endpoints.RegHealthCheckEndpoint,
	}

	r := registration.MakeRegHttpHandler(ctx, endpoint, logger)
	reg := svcdiscovery.ServiceDiscovery().Registration(*consulAddr, *consulPort, regAddr, *regPort, *svcName, logger)

	errChan := make(chan error)
	defer close(errChan)

	go func() {
		ilog.Println("Starting server at port", *regPort)
		reg.Register()
		handler := r
		errChan <- http.ListenAndServe(net.JoinHostPort(regAddr, *regPort), handler)
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
