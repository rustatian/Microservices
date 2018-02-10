package svcdiscovery

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	//"github.com/leonelquinteros/gorand"
	"net"
	"os"
	"strconv"
	"sync"
)

type Discovery interface {
	Registration(consulAddr, consulPort, svcAddress, svcPort, svcName string, logger log.Logger) (registrar sd.Registrar)
	Find(consulAddress, serviceName, tag *string) (address string, e error)
}

type serviceDiscovery struct{}

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

func client(consulAddr *string) *consulsd.Client {
	conf := api.DefaultConfig()
	conf.Address = *consulAddr
	consClnt, err := api.NewClient(conf)

	if err != nil {
		panic(err)
	}

	client := consulsd.NewClient(consClnt)
	return &client
}

func (s *serviceDiscovery) Registration(consulAddr, consulPort, svcAddress, svcPort, svcName string, logger log.Logger) (registrar sd.Registrar) {

	consulConfig := api.DefaultConfig()
	if len(consulAddr) > 0 {
		consulConfig.Address = net.JoinHostPort(consulAddr, consulPort)
	}

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Log("err", err)
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
	uuid := "15"

	asr := api.AgentServiceRegistration{
		ID:      uuid,
		Name:    svcName,
		Address: svcAddress,
		Port:    port,
		Tags:    []string{svcName, "Adexin"},
		Check:   &check,
	}

	return consulsd.NewRegistrar(consulsd.NewClient(consulClient), &asr, logger)

}

func (s *serviceDiscovery) Find(consulAddress, serviceName, tag *string) (address string, e error) {
	srventry, _, err := (*client(consulAddress)).Service(*serviceName, *tag, true, &api.QueryOptions{})
	if err != nil {
		fmt.Println(err)
	}

	if len(srventry) != 0 {
		return "http://" + srventry[0].Service.Address + ":" + strconv.Itoa(srventry[0].Service.Port), nil
	}

	return "", fmt.Errorf("error: no connected services")
}
