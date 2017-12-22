package vault

import (
	"math/rand"
	"os"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
)

func Register(consulAddr, consulPort, vaultAddress, vaultPort string, logger log.Logger) (registar sd.Registrar) {

	var client consulsd.Client = consClient(logger, consulAddr, consulPort)

	check := api.AgentServiceCheck{
		HTTP:     "http://" + vaultAddress + vaultPort + "/" + "health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	port, _ := strconv.Atoi(vaultPort[1:]) // remove :10000 -> 10000
	num := rand.Intn(500)                  // to make service ID unique

	asr := api.AgentServiceRegistration{
		ID:      "vaultsvc" + strconv.Itoa(num), //unique service ID
		Name:    "vaultsvc",
		Address: vaultAddress,
		Port:    port,
		Tags:    []string{"vaultsvc", "Adexin"},
		Check:   &check,
	}
	return consulsd.NewRegistrar(client, &asr, logger)
}

func consClient(logger log.Logger, consulAddr, consulPort string) consulsd.Client {
	consulConfig := api.DefaultConfig()
	if len(consulAddr) > 0 {
		consulConfig.Address = consulAddr + consulPort
	}

	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)

	}
	return consulsd.NewClient(consulClient)
}
