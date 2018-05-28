package svcdiscovery

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"

	consul "github.com/hashicorp/consul/api"
)

//Client provides an interface for getting data out of Consul
type Client interface {
	// Get a FindService from consul
	FindService(string, string, bool, *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error)

	// Register a service with local agent
	RegisterViaHttp(string, string, int) (string, error)

	// Deregister a service with local agent
	DeRegister(string) error
}

type client struct {
	consul *consul.Client
}

//NewConsul returns a Client interface for given consul address
func NewConsulHttpClient(addr string) (Client, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	client := client{
		consul: c,
	}

	return &client, nil
}

// Register a service with consul local agent
func (c *client) RegisterViaHttp(name, address string, port int) (string, error) {
	prt := strconv.Itoa(port)
	check := api.AgentServiceCheck{
		HTTP:     "http://" + address + ":" + prt + "/" + "health",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	Uuid := uuid.New().String()

	asr := &api.AgentServiceRegistration{
		ID:      Uuid,
		Name:    name,
		Address: address,
		Port:    port,
		Tags:    []string{name},
		Check:   &check,
	}

	return Uuid, c.consul.Agent().ServiceRegister(asr)
}

// DeRegister a service with consul local agent
func (c *client) DeRegister(id string) error {
	return c.consul.Agent().ServiceDeregister(id)
}

// FindService return a service
func (c *client) FindService(serviceName, tag string, passingOnly bool, q *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	addrs, meta, err := c.consul.Health().Service(serviceName, tag, passingOnly, q)
	if len(addrs) == 0 && err == nil {
		return nil, nil, fmt.Errorf("service ( %s ) was not found", serviceName)
	}
	if err != nil {
		return nil, nil, err
	}
	return addrs, meta, nil
}
