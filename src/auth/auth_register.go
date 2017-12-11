package auth

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func Register(consulAddr, consulPort, authAddress, authPort, serviceName string, logger log.Logger) (registar sd.Registrar) {

	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		if len(consulAddr) > 0 {
			consulConfig.Address = consulAddr + consulPort
		}
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	//client := consulsd.NewClient(ConsulClient(consulAddress, consulPort, logger))
	rand.Seed(time.Now().UTC().UnixNano())
	check := api.AgentServiceCheck{
		HTTP:     "http://" + authAddress + authPort + "/" + serviceName + "/" + "health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
		DockerContainerID: "eced4d59afd5085f61f017f874130bfec111fb4af172ca48904ed404317c36c0",

	}

	port, _ := strconv.Atoi(authPort)
	num := rand.Intn(100) // to make service ID unique
	asr := api.AgentServiceRegistration{
		ID:      "auth" + strconv.Itoa(num), //unique service ID
		Name:    "auth",
		Address: authAddress,
		Port:    port,
		Tags:    []string{"auth", "Adexin"},
		Check:   &check,
	}
	registar = consulsd.NewRegistrar(client, &asr, logger)
	return
}

////retrieve consul api client for make consulsd client or KV
//func ConsulClient(consulAddress string, consulPort string, logger log.Logger) *api.Client {
//	// Service discovery domain. In this example we use Consul.
//	consulConfig := api.DefaultConfig()
//	consulConfig.Address = net.JoinHostPort(consulAddress, consulPort)
//	consulClient, err := api.NewClient(consulConfig)
//	if err != nil {
//		logger.Log("err", err)
//		os.Exit(1)
//	}
//	return consulClient
//}