package registration

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	consulapi "github.com/hashicorp/consul/api"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func ConsClient(consulAddr *string) consulsd.Client {
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		if len(*consulAddr) > 0 {
			consulConfig.Address = *consulAddr
		}
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}
	return client
}

func ServiceD(service, tag string) (address string, e error) {

	clnt := ConsClient(&consAddr)
	srventry, _, err := clnt.Service(service, tag, true, &consulapi.QueryOptions{})

	if len(srventry) == 0 && err == nil {
		return "", fmt.Errorf("service ( %s ) was not found", service)
	}

	if err != nil {
		return "", err
	}

	addrs := srventry[0].Node.Address + ":" + strconv.Itoa(srventry[0].Service.Port)

	return addrs, nil
}

func Register(consulAddr, consulPort, authAddress, authPort string, logger log.Logger) (registrar sd.Registrar) {
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

	rand.Seed(time.Now().UTC().UnixNano())
	check := api.AgentServiceCheck{
		HTTP:     "http://" + authAddress + authPort + "/" + "health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	port, _ := strconv.Atoi(authPort)
	num := rand.Intn(100) // to make service ID unique
	asr := api.AgentServiceRegistration{
		ID:      "regsvc" + strconv.Itoa(num), //unique service ID
		Name:    "regsvc",
		Address: authAddress,
		Port:    port,
		Tags:    []string{"regsvc", "Adexin"},
		Check:   &check,
	}
	registrar = consulsd.NewRegistrar(client, &asr, logger)
	return
}
