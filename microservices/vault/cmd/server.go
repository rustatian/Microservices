package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"github.com/ValeryPiashchynski/TaskManager/microservices/proto/vault"
	"github.com/ValeryPiashchynski/TaskManager/microservices/vault"
	"github.com/ValeryPiashchynski/TaskManager/microservices/vault/application"
	"github.com/ValeryPiashchynski/TaskManager/svcdiscovery"
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

	reg := svcdiscovery.ServiceDiscovery().RegistrationViaHTTP(*consulAddr, *consulPort, vaultAddr, *vaultPort, *svcName, logg)
	defer reg.Deregister()

	vs := vault.NewVaultService(application.NewPasswordChecker())
	vs = vault.NewLoggingService(logg, vs)
	vsEndpoints := vault.NewVaultEndpoints(vs, *logg, stdopentracing.GlobalTracer())

	// Interrupt handler.
	c := make(chan os.Signal)
	r := vault.MakeVaultHttpHandler(vsEndpoints, *logg)
	srv := &http.Server{
		Handler:      r,
		Addr:         vaultAddr + ":" + *vaultPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	errCh := make(chan error)
	go func() {
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errCh <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		reg.Register()
		logg.WithFields(logrus.Fields{
			"Transport":  "HTTP",
			"Endpoint: ": net.JoinHostPort(vaultAddr, *vaultPort),
		}).Info("Server started")

		//Custom server with logrus
		errCh <- srv.ListenAndServe()
	}()

	go func() {
		listener, err := net.Listen("tcp", ":8081")
		if err != nil {
			errCh <- err
			return
		}

		handler := vault.MakeVaultGrpcHandler(vs)
		gRPCServer := grpc.NewServer()
		pb_vault.RegisterVaultServer(gRPCServer, handler)
		logg.WithFields(logrus.Fields{
			"Transport":  "TCP",
			"Endpoint: ": net.JoinHostPort("localhost", ":8081"),
		}).Info("Server started")

		errCh <- gRPCServer.Serve(listener)

	}()

	logg.WithFields(logrus.Fields{
		"exit": <-errCh,
	}).Info("Server stopped")
	time.Sleep(time.Second * 1)

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
