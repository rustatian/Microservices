package main

import (
	consulsd "github.com/go-kit/kit/sd/consul"
	"sync"
	"github.com/hashicorp/consul/api"
	"strconv"
)

//TODO viper config
func init() {

}

type Discovery interface {
	Find(consulAddress, serviceName, tag *string) (address string, e error)
}

type serviceDiscovery struct {}

var instance *serviceDiscovery
var once sync.Once


func GetServiceDiscovery() Discovery {
	once.Do(func() {
		instance = &serviceDiscovery{}
	})
	return instance
}



func client(consulAddr *string) *consulsd.Client{
	conf := api.DefaultConfig()
	conf.Address = *consulAddr
	consClnt, err := api.NewClient(conf)

	if err != nil {
		panic(err)
	}

	client := consulsd.NewClient(consClnt)
	return &client
}

func(s *serviceDiscovery) Find(consulAddress, serviceName, tag *string) (address string, e error) {
	srventry, _, err := (*client(consulAddress)).Service(*serviceName, *tag, true, &api.QueryOptions{})
	if err != nil {
		panic(err)
	}

	addrs := srventry[0].Node.Address + ":" + strconv.Itoa(srventry[0].Service.Port)

	return addrs, nil
}

//func ConsClient(consulAddr *string) consulsd.Client {
//	var client consulsd.Client
//	{
//		consulConfig := api.DefaultConfig()
//		if len(*consulAddr) > 0 {
//			consulConfig.Address = *consulAddr
//		}
//		consulClient, err := api.NewClient(consulConfig)
//		if err != nil {
//			os.Exit(1)
//		}
//		client = consulsd.NewClient(consulClient)
//	}
//	return client
//}




//func ServiceD(service, tag string) (address string, e error) {
//
//	clnt := ConsClient(&consAddr)
//	srventry, _, err := clnt.Service(service, tag, true, &consulapi.QueryOptions{})
//
//	if len(srventry) == 0 && err == nil {
//		return "", fmt.Errorf("service ( %s ) was not found", service)
//	}
//
//	if err != nil {
//		return "", err
//	}
//
//	addrs := srventry[0].Node.Address + ":" +strconv.Itoa(srventry[0].Service.Port)
//
//	return addrs, nil
//}

