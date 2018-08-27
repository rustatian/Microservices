package servicediscovery

import (
	"fmt"

	"github.com/hashicorp/consul/api"

	"net"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/inturnco/go-sdk/helpers"
	consul "github.com/hashicorp/consul/api"
)

//Client provides an interface for getting data out of Consul
type Client interface {
	// Get a FindService from consul
	FindService(string, string, bool, *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error)

	// Register a service with local agent
	RegisterViaHTTP(string, string, string) (string, error)

	// Deregister a service with local agent
	DeRegister(string) error
}

type client struct {
	consul *consul.Client
}

// NewDefaultConsulHTTPClient returns a Client interface for given consul address
// Default consul http client returns a default configuration for the client with provided address
// By default it will use 15 second http client timeout and also default consul client config
func NewDefaultConsulHTTPClient(addr string) (Client, error) {
	config := consul.DefaultConfig()
	cln := &http.Client{
		Timeout: time.Duration(time.Second * 10),
	}

	config.Address = addr
	config.HttpClient = cln

	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	client := client{
		consul: c,
	}

	return &client, nil
}

// RegisterViaHTTP registers a service with consul local agent
func (c *client) RegisterViaHTTP(name, address, port string) (string, error) {
	check := api.AgentServiceCheck{
		HTTP:     "http://" + net.JoinHostPort(address, port) + "/healthcheck",
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	// Generate UUID
	UUID, err := helpers.GenerateUUID()
	if err != nil {
		return "", err
	}

	// Convert port to int to use in AgentServiceRegistration
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return "", err
	}

	asr := &api.AgentServiceRegistration{
		ID:      UUID,
		Name:    name,
		Address: address,
		Port:    portInt,
		Tags:    []string{name},
		Check:   &check,
	}

	return UUID, c.consul.Agent().ServiceRegister(asr)
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
