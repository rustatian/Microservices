package registration

import (
	consulsd "github.com/go-kit/kit/sd/consul"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api"
	"os"
	"fmt"
	"strconv"
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

	client := ConsClient(&consAddr)
	srventry, _, err := client.Service(service, tag, true, &consulapi.QueryOptions{})

	if len(srventry) == 0 && err == nil {
		return "", fmt.Errorf("service ( %s ) was not found", service)
	}

	if err != nil {
		return "", err
	}

	addrs := srventry[0].Node.Address + ":" +strconv.Itoa(srventry[0].Service.Port)

	return addrs, nil
}