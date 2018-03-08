package svcdiscovery

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/go-kit/kit/sd"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

type Discovery interface {
	RegistrationViaHTTP(consulAddr, consulPort, svcAddress, svcPort, svcName string, logger *logrus.Logger) (registrar sd.Registrar)
	RegistrationViaTCP(consulAddr, consulPort, svcAddress, svcPort, svcName string, logger *logrus.Logger) (registrar sd.Registrar)
	Find(consulAddress, serviceName, tag *string) (address string, e error)
}

type serviceDiscovery struct {
}

var (
	instance *serviceDiscovery
	once     sync.Once
)

func ServiceDiscovery() Discovery {
	once.Do(func() {
		instance = &serviceDiscovery{}
	})
	return instance
}

func consuldClient(consulAddr *string) *Client {
	conf := api.DefaultConfig()
	conf.Address = *consulAddr
	consClnt, err := api.NewClient(conf)

	if err != nil {
		panic(err)
	}

	client := NewClient(consClnt)
	return &client
}

func (s *serviceDiscovery) RegistrationViaHTTP(consulAddr, consulPort, svcAddress, svcPort, svcName string, logger *logrus.Logger) (registrar sd.Registrar) {

	consulConfig := api.DefaultConfig()
	if len(consulAddr) > 0 {
		consulConfig.Address = net.JoinHostPort(consulAddr, consulPort)
	}

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("RegistrationViaHTTP error")
		os.Exit(1)
	}

	check := api.AgentServiceCheck{
		HTTP:     "http://" + net.JoinHostPort(svcAddress, svcPort) + "/" + "health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	port, err := strconv.Atoi(svcPort)
	if err != nil {

	}
	Uuid := uuid.New().String()

	asr := api.AgentServiceRegistration{
		ID:      Uuid,
		Name:    svcName,
		Address: svcAddress,
		Port:    port,
		Tags:    []string{svcName, "ValeryPiashchynski"},
		Check:   &check,
	}

	return NewRegistrar(NewClient(consulClient), &asr, logger)

}

func (s *serviceDiscovery) RegistrationViaTCP(consulAddr, consulPort, svcAddress, svcPort, svcName string, logger *logrus.Logger) (registrar sd.Registrar) {

	consulConfig := api.DefaultConfig()
	if len(consulAddr) > 0 {
		consulConfig.Address = net.JoinHostPort(consulAddr, consulPort)
	}

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("RegistrationViaHTTP error")
		os.Exit(1)
	}

	check := api.AgentServiceCheck{
		TCP:      net.JoinHostPort(svcAddress, svcPort),
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	port, err := strconv.Atoi(svcPort)
	if err != nil {

	}

	Uuid := uuid.New().String()
	asr := api.AgentServiceRegistration{
		ID:      Uuid,
		Name:    svcName,
		Address: svcAddress,
		Port:    port,
		Tags:    []string{svcName, "ValeryPiashchynski"},
		Check:   &check,
	}

	return NewRegistrar(NewClient(consulClient), &asr, logger)

}

func (s *serviceDiscovery) Find(consulAddress, serviceName, tag *string) (address string, e error) {
	srventry, _, err := (*consuldClient(consulAddress)).Service(*serviceName, *tag, true, &api.QueryOptions{})
	if err != nil {
		fmt.Println(err)
	}

	if len(srventry) != 0 {
		return "http://" + srventry[0].Service.Address + ":" + strconv.Itoa(srventry[0].Service.Port), nil
	}

	return "", fmt.Errorf("error: no connected services")
}
