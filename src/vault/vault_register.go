package vault

import (
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"strconv"
	"math/rand"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/log"
)

func Register(vaultAddress, vaultPort, serviceName string, client consulsd.Client, logger log.Logger) (registar sd.Registrar) {
	check := api.AgentServiceCheck{
		HTTP:     "http://" + vaultAddress + vaultPort + "/" + serviceName + "/" + "health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	port, _ := strconv.Atoi(vaultPort)
	num := rand.Intn(100) // to make service ID unique
	asr := api.AgentServiceRegistration{
		ID:      serviceName + strconv.Itoa(num), //unique service ID
		Name:    serviceName,
		Address: vaultAddress,
		Port:    port,
		Tags:    []string{"vaultsvc", "Adexin"},
		Check:   &check,
	}

	return consulsd.NewRegistrar(client, &asr, logger)
}
