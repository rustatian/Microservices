package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "Microservices/svcdiscovery"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"

	gonats "github.com/nats-io/go-nats"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"Microservices/microservices/proto/vault"
	"Microservices/microservices/vault"
	"Microservices/microservices/vault/application"
)

func main() {
	var (
		consulAddr    = flag.String("consul.addr", "localhost", "consul address")
		consulPort    = flag.Int("consul.port", 8500, "consul port")
		vaultHttpPort = flag.Int("vault.http", 10000, "vault port")
		vaultTcpPort  = flag.String("vault.tcp", ":8081", "vault tcp port")
		svcName       = flag.String("service.name", "vaultsvc", "Vault service name")
	)

	vaultAddr, _ := externalIP()

	flag.Parse()

	logg := logrus.New()
	logg.Out = os.Stdout

	consPort := strconv.Itoa(*consulPort)
	reg, err := svcdiscovery.NewConsulHttpClient(net.JoinHostPort(*consulAddr, consPort))
	if err != nil {
		println(err.Error())
	}
	uuid, err := reg.RegisterViaHttp(*svcName, vaultAddr, *vaultHttpPort)
	//reg := svcdiscovery.ServiceDiscovery().RegistrationViaHTTP(*consulAddr, *consulPort, vaultAddr, *vaultHttpPort, *svcName, logg)
	defer reg.DeRegister(uuid)

	// variables of application level
	hasher := application.NewBcryptHasher()
	validator := application.NewBcryptValidator()
	healthChecker := application.NewHttpHealthChecker()

	vs := vault.NewVaultService(hasher, validator, healthChecker)
	//vs = vault.NewLoggingService(logg, vs)
	vsEndpoints := vault.NewVaultEndpoints(vs, *logg, stdopentracing.GlobalTracer())

	// Interrupt handler.
	c := make(chan os.Signal)
	r := vault.MakeVaultHttpHandler(vsEndpoints, *logg)
	srv := &http.Server{
		Handler:      r,
		Addr:         vaultAddr + ":" + "10000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	errCh := make(chan error)

	go func() {
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		errCh <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		if err != nil {
			println(err.Error())
		}
		logg.WithFields(logrus.Fields{
			"Transport":  "HTTP",
			"Endpoint: ": net.JoinHostPort(vaultAddr, "10000"),
		}).Info("Server started")

		// Custom server with logrus
		errCh <- srv.ListenAndServe()
	}()

	go func() {
		listener, err := net.Listen("tcp", *vaultTcpPort)
		if err != nil {
			errCh <- err
			return
		}

		handler := vault.MakeVaultGrpcHandler(vs)
		gRPCServer := grpc.NewServer()
		pb_vault.RegisterVaultServer(gRPCServer, handler)
		logg.WithFields(logrus.Fields{
			"Transport":  "TCP",
			"Endpoint: ": net.JoinHostPort(vaultAddr, *vaultTcpPort),
		}).Info("Server started")

		errCh <- gRPCServer.Serve(listener)

	}()

	go func() {
		// nc, er := gonats.Connect(gonats.DefaultURL)
		// if er != nil {
		// 	errCh <- er
		// }
		nc, err := gonats.Connect(gonats.DefaultURL)
		gonats.DisconnectHandler(func(_ *gonats.Conn) {
			fmt.Printf("Got disconnected!\n")
		})
		gonats.ReconnectHandler(func(_ *gonats.Conn) {
			fmt.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
		})
		gonats.ClosedHandler(func(_ *gonats.Conn) {
			fmt.Printf("Connection closed. Reason: %q\n", nc.LastError())
		})

		if err != nil {
			println(err)
			return
		}

		vault.StartVaultNatsHandler(vsEndpoints, *logg, nc)
	}()

	go func() {
		conn, err := amqp.Dial("amqp://guest@localhost:5672")
		if err != nil {
			errCh <- err
		}

		ch, err := conn.Channel()
		if err != nil {
			errCh <- err
		}

		q, err := ch.QueueDeclare(
			"hash",
			false,
			false,
			false,
			false,
			nil,
		)

		msgs, err := ch.Consume(
			q.Name,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			errCh <- err
		}

		nc, err := gonats.Connect(gonats.DefaultURL)
		if err != nil {
			errCh <- err
		}

		for {
			select {
			case mg, ok := <-msgs:
				if !ok {
					return
				}

				log.Printf("Received a message: %s", mg.Body)
				nc.Publish("hash", mg.Body)

				aa, _, _ := reg.FindService("vaultsvc", "vaultsvc", true, nil)
				for _, a := range aa {
					println(a.Checks.AggregatedStatus())
					addr := a.Service.Address
					println(addr)
				}

			case er := <-errCh:
				println(er)
				return
			}
		}
	}()

	logg.WithFields(logrus.Fields{
		"exit": <-errCh,
	}).Info("Server stopped")
	// time.Sleep(time.Second * 5)

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
