package vault

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/leonelquinteros/gorand"
	"net"
	"os"
	"strconv"
)

func Register(consulAddr, consulPort, vaultAddress, vaultPort string, logger log.Logger) (registar sd.Registrar) {

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
		HTTP:     "http://" + net.JoinHostPort(vaultAddress, vaultPort) + "/" + "health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	port, _ := strconv.Atoi(vaultPort) // remove :10000 -> 10000
	uuid, _ := gorand.UUID()

	asr := api.AgentServiceRegistration{
		ID:      uuid,
		Name:    "vaultsvc",
		Address: vaultAddress,
		Port:    port,
		Tags:    []string{"vaultsvc", "Adexin"},
		Check:   &check,
	}

	return consulsd.NewRegistrar(consulsd.NewClient(consulClient), &asr, logger)
}
