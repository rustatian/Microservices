package authorization

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

func Register(consulAddr, consulPort, authAddress, authPort string, logger log.Logger) (registrar sd.Registrar) {

	consulConfig := api.DefaultConfig()

	if len(consulAddr) > 0 {
		consulConfig.Address = consulAddr + consulPort
	}
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
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
		ID:      "auth" + strconv.Itoa(num), //unique service ID
		Name:    "auth",
		Address: authAddress,
		Port:    port,
		Tags:    []string{"auth", "Adexin"},
		Check:   &check,
	}
	registrar = consulsd.NewRegistrar(consulsd.NewClient(consulClient), &asr, logger)
	return
}

