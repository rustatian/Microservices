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

	"github.com/ValeryPiashchynski/TaskManager/microservices/vault"
	"github.com/ValeryPiashchynski/TaskManager/svcdiscovery"
	"github.com/go-kit/kit/log"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		consulAddr = flag.String("consul.addr", "localhost", "consul address")
		consulPort = flag.String("consul.port", "8500", "consul port")
		vaultPort  = flag.String("vault.port", "10000", "vault port")
		svcName    = flag.String("service.name", "vaultsvc", "Vault service name")
	)

	vaultAddr := "localhost"
	//vaultAddr, _ := externalIP()

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

	svc := vault.NewVaultService()

	tracer := stdopentracing.GlobalTracer()
	reg := svcdiscovery.ServiceDiscovery().RegistrationViaHTTP(*consulAddr, *consulPort, vaultAddr, *vaultPort, *svcName, logger)
	defer reg.Deregister()

	endpoints := vault.NewEndpoints(svc, logger, tracer)

	errCh := make(chan error)
	// Interrupt handler.
	c := make(chan os.Signal)

	//Error handler
	go func() {
		//logger.Log("nats error:", <-errCh)
		stdlog.Fatal(<-errCh)
	}()

	//r := vault.MakeVaultNatsHandler(endpoints, logger, "nats://172.24.231.70:4222", errCh)
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
