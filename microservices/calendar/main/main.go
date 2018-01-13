package main

import (
	"TaskManager/microservices/calendar"
	"TaskManager/svcdiscovery"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
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
		tCalPort   = flag.String("tcal.port", "10003", "task calendar port")
		svcName    = flag.String("service.name", "tcalsvc", "Task calendar service name")
	)

	tCalAddr, _ := externalIP()

	flag.Parse()
	ctx := context.Background()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	svc := calendar.NewService()

	tracer := stdopentracing.GlobalTracer()
	reg := svcdiscovery.ServiceDiscovery().Registration(*consulAddr, *consulPort, tCalAddr, *tCalPort, *svcName, logger)
	defer reg.Deregister()

	endpoints := calendar.NewEndpoints(svc, logger, tracer)
	r := calendar.MakeTcHttpHandler(ctx, endpoints, logger)

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
		logger.Log("transport", "HTTP", "addr", ":"+*tCalPort)
		errc <- http.ListenAndServe(":"+*tCalPort, r)
	}()

	logger.Log("exit", <-errc)
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
